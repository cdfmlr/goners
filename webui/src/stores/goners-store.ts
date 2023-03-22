import { defineStore } from 'pinia';
import { Device, Addr, Packet, Layer } from '../model/model';
import * as api from '../api/http';

import { Notify } from 'quasar';

export const useGonersStore = defineStore('goners', {
  state: () => ({
    devices: [] as Device[],
    selectedDevice: null as Device | null,
    bpfFilter: '' as string,

    // pcap session
    started: false as boolean,
    paused: false as boolean,

    pcapSessionID: null as string | null,
    pcapWS: null as WebSocket | null,

    // fetched packets
    packets: [] as Packet[],
    selectedPacket: null as Packet | null,
    selectedLayer: null as Layer | null,
  }),
  getters: {
    getDevices(): Device[] {
      return this.devices;
    },
    isDeviceSelected(): boolean {
      return !!this.selectedDevice;
    },
  },
  actions: {
    async fetchDevices() {
      try {
        this.devices = await api.getDevices();
        console.log('devices', this.devices);

        if (this.devices.length > 0 && this.selectedDevice === null) {
          // this.selectedDevice = this.devices[0];
        }
      } catch (e) {
        Notify.create({
          type: 'negative',
          message: `Error fetching devices: ${e}`,
        });
      }
    },
    async startPcap() {
      if (!this.selectedDevice?.name) {
        Notify.create({
          type: 'negative',
          message: 'No device selected',
        });
        return;
      }

      try {
        this.pcapSessionID = await api.startPcap({
          device: this.selectedDevice?.name || '',
          filter: this.bpfFilter,
        });
      } catch (e) {
        Notify.create({
          type: 'negative',
          message: `Error starting pcap: ${e}`,
        });
        return;
      }

      if (!this.pcapSessionID) {
        Notify.create({
          type: 'negative',
          message: 'No session ID',
        });
        return;
      }
      try {
        this.pcapWS = await api.getPackets(this.pcapSessionID!);
      } catch (e) {
        Notify.create({
          type: 'negative',
          message: `Error getting packets: ${e}`,
        });
        return;
      }

      this._receivePackets();

      this.pcapWS.onclose = (e) => {
        Notify.create({
          type: 'negative',
          message: `PCAP session closed: ${e.reason}`,
        });
        this.pcapSessionID = null;
        this.pcapWS = null;
      };

      this.started = true;
      this.paused = false;
    },
    async stopPcap() {
      if (!this.pcapSessionID) {
        throw new Error('No session ID');
      }
      await api.stopPcap(this.pcapSessionID);
      this.pcapSessionID = null;
      this.pcapWS?.close();
      this.pcapWS = null;
      this.started = false;
      this.paused = false;
    },
    async pausePcap() {
      if (!this.pcapSessionID) {
        throw new Error('No session ID');
      }

      if (this.paused) {
        this._receivePackets();
        this.paused = false;
      } else {
        this._ignorePackets();
        this.paused = true;
      }
    },
    clearPackets() {
      this.packets = [];
    },
    savePackets() {
      if (this.packets.length === 0) {
        Notify.create({
          type: 'negative',
          message: 'No packets to save',
        });
        return;
      }

      const blob = new Blob([JSON.stringify(this.packets)], {
        type: 'application/json',
      });
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');

      // download attribute is not supported in Safari
      // https://stackoverflow.com/questions/3665115/create-a-file-in-memory-for-user-to-download-not-through-server
      link.setAttribute('href', url);
      link.setAttribute('download', 'goners-packets.json');
      link.style.visibility = 'hidden';
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
    },
    _receivePackets() {
      if (!this.pcapWS) {
        throw new Error('No pcap websocket');
      }
      this.pcapWS.onmessage = (e) => {
        const packet = JSON.parse(e.data) as Packet;
        this.packets.push(packet);
      };
    },
    _ignorePackets() {
      if (!this.pcapWS) {
        throw new Error('No pcap websocket');
      }
      this.pcapWS.onmessage = () => {
        // do nothing
      };
    },
  },
});
