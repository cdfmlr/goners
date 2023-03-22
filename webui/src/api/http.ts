// This file contains the functions that interact with the API endpoints.
//
// goners http api:
//
// devicse:
//   GET  /devices: lookup devices
// pcap:
//   POST   /pcap:  start a capturing
//   DELETE /pcap:  stop a capturing
//   WS     /pcap/{sessionID}: get packets
//
// Using fetch API.

import { Device, PcapConfig } from '../model/model';

const baseURL = 'http://127.0.0.1:9801';

// Get the devices from the API.
export function getDevices(): Promise<Device[]> {
  return fetch(`${baseURL}/devices`).then((res) => {
    if (res.ok) {
      return res.json();
    } else {
      throw new Error(`Error getting devices: ${res.status} ${res.statusText}`);
    }
  });
}

// Start a PCAP session.
export function startPcap(config: PcapConfig): Promise<string> {
  return fetch(`${baseURL}/pcap`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(config),
  })
    .then((res) => {
      if (res.ok) {
        return res.json();
      } else {
        throw new Error(`Error starting pcap: ${res.status} ${res.statusText}`);
      }
    })
    .then((json) => json.session_id);
}

// Stop a PCAP session.
export function stopPcap(sessionID: string): Promise<string> {
  return fetch(`${baseURL}/pcap`, {
    method: 'DELETE',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ session_id: sessionID }),
  })
    .then((res) => {
      if (res.ok) {
        return res.json();
      } else {
        throw new Error(`Error stopping pcap: ${res.status} ${res.statusText}`);
      }
    })
    .then((json) => json.deleted_session_id);
}

// Get the packets from the API.
export function getPackets(sessionID: string): Promise<WebSocket> {
  return new Promise((resolve, reject) => {
    const ws = new WebSocket(
      `${baseURL}/pcap/${sessionID}`.replace('http://', 'ws://')
    );
    ws.onopen = () => resolve(ws);
    ws.onerror = (e) => reject(e);
  });
}
