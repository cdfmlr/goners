<template>
  <q-layout view="lHh Lpr lFf">
    <q-header elevated>
      <q-toolbar class="row wrap justify-around">
        <q-btn
          flat
          dense
          round
          icon="menu"
          aria-label="Menu"
          @click="toggleLeftDrawer"
        />

        <q-toolbar-title>
          Goners
          <q-badge v-if="store.packets" align="middle">
            ({{ store.packets.length }})
          </q-badge>
        </q-toolbar-title>

        <div class="row wrap justify-around">
          <q-select
            class="haeader-input"
            dark
            standout
            v-model="store.selectedDevice"
            :options="devices"
            label="device"
            :display-value="`<pre style='margin:0'>${displaySelectedDevice()}</pre>`"
            :display-value-html="true"
          >
            <template v-slot:option="scope">
              <q-item v-bind="scope.itemProps">
                <q-item-section>
                  <q-item-label>
                    {{ scope.opt.index }} {{ scope.opt.name }}
                  </q-item-label>

                  <q-item-label caption>
                    <pre>{{
                      scope.opt.addrs
                        .map((addr: Addr) => formatAddr(addr))
                        .join('\n')
                    }}</pre>
                  </q-item-label>
                </q-item-section>
              </q-item>
            </template>
          </q-select>

          <q-input
            class="haeader-input"
            dark
            standout
            v-model="store.bpfFilter"
            label="filter"
            style="min-width: 20vw"
            input-style="font-family: monospace;"
          />

          <!-- <q-space /> -->

          <q-btn
            :label="'start' + (store.started ? 'ed' : '')"
            :disable="!store.isDeviceSelected || store.started"
            @click="store.startPcap()"
            flat
            stack
            icon="play_arrow"
          />
          <q-btn
            label="pause"
            :disable="!store.started"
            @click="store.pausePcap()"
            flat
            stack
            icon="pause"
          />
          <q-btn
            label="stop"
            :disable="!store.started"
            @click="store.stopPcap()"
            flat
            stack
            icon="stop"
          />
          <q-btn
            label="clear"
            :disable="store.packets.length === 0"
            flat
            stack
            @click="store.clearPackets()"
            icon="clear"
          />
          <q-btn
            label="save"
            :disable="store.packets.length === 0"
            flat
            stack
            @click="store.savePackets()"
            icon="save"
          />
        </div>

        <!-- <div>webui v0.0.0</div> -->
      </q-toolbar>
    </q-header>

    <q-drawer v-model="leftDrawerOpen" bordered>
      <q-img
        class="absolute-top"
        src="https://cdn.quasar.dev/img/material.png"
        style="height: 150px"
      >
        <div class="absolute-bottom bg-transparent">
          <q-avatar size="56px" class="q-mb-sm" color="red" text-color="white">
            G
          </q-avatar>
          <div class="text-weight-bold">goners</div>
          <div style="font-size: x-small">
            Goner's Oafish Network Explorer / Reliable Sniffer
          </div>
        </div>
      </q-img>
      <q-scroll-area
        style="
          height: calc(100% - 150px);
          margin-top: 150px;
          border-right: 1px solid #ddd;
        "
      >
        <q-list>
          <!-- <q-item-label header> About </q-item-label> -->

          <EssentialLink
            v-for="link in essentialLinks"
            :key="link.title"
            v-bind="link"
          />
        </q-list>
      </q-scroll-area>
      <div
        class="absolute-bottom"
        style="margin: 8px; text-align: center; color: grey; font-size: small"
      >
        <div>v0.0.0</div>
        <div style="font-size: x-small">
          Copyright &copy; 2023 CDFMLR All Rights Reserved.
        </div>
      </div>
    </q-drawer>

    <q-page-container>
      <!-- <router-view /> -->
      <PacketsView :packets="store.packets" @select="selectPacket" />
      <LayersView :packet="store.selectedPacket" @select-layer="selectLayer" />
      <DumpView :layer="store.selectedLayer" />
    </q-page-container>
  </q-layout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import EssentialLink, {
  EssentialLinkProps,
} from 'components/EssentialLink.vue';
import PacketsView from 'components/PacketsView.vue';
import LayersView from 'components/LayersView.vue';
import { useGonersStore } from '../stores/goners-store';
import { Addr, formatAddr, Layer, Packet } from 'src/model/model';
import DumpView from 'src/components/DumpView.vue';

const store = useGonersStore();

const essentialLinks: EssentialLinkProps[] = [
  {
    title: 'Backend',
    caption: 'Golang with gin, cli and gopacket',
    icon: 'dns',
    link: 'https://github.com/cdfmlr/goners',
  },
  {
    title: 'Frontend',
    caption: 'Quasar with Vue3, TypeScript, Pinia and Vite',
    icon: 'desktop_windows',
    link: 'https://github.com/cdfmlr/goners',
  },
  {
    title: 'GitHub',
    caption: 'github.com/cdfmlr/goners',
    icon: 'code',
    link: 'https://github.com/cdfmlr/goners',
  },
  {
    title: 'Developed by CDFMLR',
    // caption: '+ ChatGPT + GitHub Copilot',
    icon: 'favorite',
    link: 'https://github.com/cdfmlr',
  },
];

function displaySelectedDevice() {
  if (!store.selectedDevice) {
    return `${'undefined'.padStart(2 + 1 + 8)}`;
  }
  return `${store.selectedDevice?.index
    .toString()
    .padStart(2, '0')} ${store.selectedDevice?.name.padStart(8)}`;
}

const devices = computed(() => store.devices);
const selectedDevice = ref(store.selectedDevice);
const bpfFilter = ref(store.bpfFilter);

onMounted(() => {
  store.fetchDevices();
});

function selectPacket(index: number, packet: Packet) {
  console.log('selectPacket', index, packet);
  if (store.packets[index] !== packet) {
    console.error('selectPacket: packet mismatch. Use index.');
    packet = store.packets[index];
  }
  store.selectedPacket = packet;
}

function selectLayer(index: number, layer: Layer | undefined) {
  console.log('selectLayer', index, layer);
  if (store.selectedPacket?.layers[index] !== layer) {
    console.error('selectLayer: layer mismatch. Use index.');
    layer = store.selectedPacket?.layers[index];
  }
  store.selectedLayer = layer!;
}

const leftDrawerOpen = ref(false);

function toggleLeftDrawer() {
  leftDrawerOpen.value = !leftDrawerOpen.value;
}
</script>

<style lang="scss">
.header-input {
  width: 200px;
  margin: 8px;
  padding: 4px;
  // background-color: white;
}
</style>
