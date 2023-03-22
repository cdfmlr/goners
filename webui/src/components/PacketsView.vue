<template>
  <div>
    <div v-if="props.packets.length > 0">
      <!-- packets list -->
      <q-virtual-scroll
        type="table"
        dense
        style="max-height: 30vh"
        :items="props.packets"
        :virtual-scroll-item-size="48"
        :virtual-scroll-sticky-size-start="48"
        :virtual-scroll-sticky-size-end="32"
        v-model.number="virtualListIndex"
        @virtual-scroll="onVirtualScroll"
        ref="virtualListRef"
      >
        <template v-slot:before>
          <thead class="thead-sticky text-left">
            <tr>
              <th>Index</th>
              <th v-for="col in columns" :key="'1--' + col.name">
                {{ col.name }}
              </th>
            </tr>
          </thead>
        </template>

        <template v-slot:after>
          <tfoot class="tfoot-sticky text-left">
            <tr>
              <th>Index</th>
              <th v-for="col in columns" :key="'2--' + col.name">
                {{ col.name }}
              </th>
            </tr>
          </tfoot>
        </template>

        <template v-slot="{ item: row, index }">
          <tr
            :key="index"
            :style="`background: ${rowBgColor(index, row)};`"
            :elevated="index === selectedIndex"
            :color="index === selectedIndex ? '#fff' : ''"
            @click="onRowClick(index)"
          >
            <!-- q-chip: workaround 选中者白字，否则黑字；还带来一个有趣的特性：点击字段会有圆角矩形选中动画 -->
            <td>
              <q-chip
                style="background: #0000; margin: 0"
                dense
                :dark="isPacketSelected(index)"
              >
                #{{ index }}
              </q-chip>
            </td>
            <td v-for="col in columns" :key="index + '-' + col.name">
              <q-chip
                style="background: #0000; margin: 0"
                dense
                :dark="isPacketSelected(index)"
              >
                {{ row[col.prop] }}
              </q-chip>
            </td>
          </tr>
        </template>
      </q-virtual-scroll>
    </div>
    <div v-else>
      <q-card style="height: 30vh">
        <q-card-section>
          <q-stepper flat color="primary" animated>
            <q-step
              name="1"
              title="Select Device"
              :caption="`${
                store.selectedDevice
                  ? store.selectedDevice.name
                  : 'No device selected'
              }`"
              :done="store.isDeviceSelected"
              icon="router"
            />
            <q-step
              name="2"
              title="Set Filter"
              caption="(Optional)"
              :done="store.bpfFilter !== ''"
              icon="filter_alt"
            />
            <q-step
              name="3"
              title="Start Capture"
              :done="store.packets.length > 0"
              icon="play_arrow"
            />
          </q-stepper>
          <!-- skeleton list -->
          <q-skeleton
            v-for="i in 4"
            :key="i"
            style="margin-bottom: 8px"
            animation="blink"
            :rows="10"
            :row-height="32"
            :row-width="['100%', '100%', '100%', '100%', '100%']"
          />
        </q-card-section>
      </q-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUpdate, onMounted, ref } from 'vue';
import { Packet } from '../model/model';

// for steps
import { useGonersStore } from '../stores/goners-store';
const store = useGonersStore();

interface Props {
  packets: Packet[];
}

const props = withDefaults(defineProps<Props>(), {
  packets: () => [],
});

const emit = defineEmits(['select']);

const selectedIndex = ref(null as number | null);

const columns = [
  { name: 'timestamp', prop: 'timestamp' },
  { name: 'src', prop: 'src' },
  { name: 'dst', prop: 'dst' },
  { name: 'length', prop: 'length' },
  { name: 'capture_length', prop: 'capture_length' },
  { name: 'proto', prop: 'packet_type' },
  // { name: 'info', prop: 'info' },
];

function rowBgColor(index: number, packet: Packet) {
  if (index === selectedIndex.value) {
    return '#3574F2';
  }
  switch (packet.packet_type) {
    case 'Payload':
      const colors = ['#F2C6C2', '#F2E1C2'];
      return colors[index % colors.length];
    case 'TCP':
      return '#D9F1F2';
    case 'UDP':
      return '#CEDEF2';
    case 'ICMP':
      return '#CEF2DF';
    default:
      return '#FFFFFF';
  }
}

function onRowClick(index: number) {
  selectedIndex.value = index;
  emit('select', index, props.packets[index]);
}

function isPacketSelected(index: number) {
  return index === selectedIndex.value;
}

const virtualListRef = ref(null as any);
const virtualListIndex = ref(0 as string | number);

function onVirtualScroll(a: { index: string | number }) {
  virtualListIndex.value = a.index;
}

onMounted(() => {
  virtualListRef?.value?.scrollTo(virtualListIndex.value);
});

var lastPacketsLength = 0;

// sticky last row
onBeforeUpdate(() => {
  console.log(
    'onBeforeUpdate: ',
    virtualListIndex.value,
    props.packets.length,
    lastPacketsLength
  );
  // 5: 当前在底部才跟踪
  // (len - last): 新增的包个数
  var delta = 5 + (props.packets.length - lastPacketsLength);

  if (virtualListIndex.value >= props.packets.length - delta) {
    virtualListIndex.value = props.packets.length;
  }
  virtualListRef?.value?.scrollTo(virtualListIndex.value);
  lastPacketsLength = props.packets.length;
});
</script>

<style scoped></style>
