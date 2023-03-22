<template>
  <div>
    <div v-if="props.packet">
      <q-card style="height: 30vh" dense>
        <q-card-section>
          <q-scoll-area style="height: 30vh" dense>
            <!-- packet info -->
            <q-item dense>
              <q-item-section>
                <q-item-label>
                  <pre style="margin: 0">
PACKET {{ props.packet?.packet_type }}: {{ props.packet?.src }} -> {{
                      props.packet?.dst
                    }} @ {{ props.packet?.timestamp }}
              Length: {{ props.packet?.length }} bytes ({{
                      props.packet?.capture_length
                    }} captured)
              </pre
                  >
                </q-item-label>
                <span style="color: grey">Layers:</span>
              </q-item-section>
            </q-item>
            <!-- layers -->
            <q-item
              v-for="(layer, index) in packet?.layers"
              dense
              :key="index"
              clickable
              :style="`background: ${bgColor(index)};`"
              :dark="isLayerSelected(index)"
              @click="selectLayer(index)"
            >
              <q-item-section>
                <q-item-label>#{{ index }} </q-item-label>
                <!-- <q-item-label caption>{{ layer.fields }}</q-item-label> -->
              </q-item-section>
              <q-item-section>
                {{ layer.layer_type }}
              </q-item-section>
              <q-item-section>
                {{ layer.src }}
              </q-item-section>
              <q-item-section>
                {{ layer.src ? '->' : '' }}
              </q-item-section>
              <q-item-section>
                {{ layer.dst }}
              </q-item-section>
              <q-item-section>
                {{ layerPayloadLength(layer) }}
              </q-item-section>
              <!-- <q-item-section>
            <q-space />
          </q-item-section> -->
            </q-item>
          </q-scoll-area>
        </q-card-section>
      </q-card>
    </div>
    <div v-else>
      <q-card style="height: 30vh" dense>
        <q-card-section>
          <q-scoll-area style="height: 30vh" dense>
            <!-- packet info -->
            <!-- <q-skeleton animation="blink" type="text" height="6.5em" /> -->
            <q-stepper flat color="primary" animated>
              <q-step
                name="1"
                title="Select a packet"
                caption="Select a packet to view its information and layers."
                icon="layers"
              />
            </q-stepper>
            <!-- layers -->
            <q-skeleton
              v-for="i in 3"
              :key="i"
              style="margin-bottom: 8px"
              animation="blink"
              :rows="10"
              :row-height="32"
              :row-width="['100%', '100%', '100%', '100%', '100%']"
            />
          </q-scoll-area>
        </q-card-section>
      </q-card>
    </div>
  </div>
</template>
<script setup lang="ts">
import { Layer, Packet } from 'src/model/model';
import { computed, ref } from 'vue';

interface Props {
  packet: Packet | null;
}

const props = withDefaults(defineProps<Props>(), {
  packet: () => ({} as Packet),
});

const emit = defineEmits(['selectLayer']);

const selectedLayerIndex = ref(null as number | null);
const selectedLayer = ref(undefined as Layer | undefined);

function selectLayer(index: number) {
  selectedLayerIndex.value = index;
  selectedLayer.value = props.packet?.layers[index];
  emit('selectLayer', index, props.packet?.layers[index]);
}

function layerPayloadLength(layer: any): string {
  try {
    return `${layer.payload.length} bytes`;
  } catch (e) {
    return 'NA';
  }
}

// colors: each layer has a color: Link, Network, Transport, Application
const bgColors = ['#ffffff', '#efefef', '#ffffff', '#efefef'];

function bgColor(index: number): string {
  if (isLayerSelected(index)) {
    return '#3574F2';
  }
  return bgColors[index % bgColors.length];
}

function isLayerSelected(index: number): boolean {
  return (
    index === selectedLayerIndex.value &&
    props.packet?.layers[index] === selectedLayer.value
  );
}
</script>
