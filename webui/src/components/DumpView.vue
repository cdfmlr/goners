<!-- show selected layer's fields (left) & dump (right) -->
<template>
  <div>
    <div class="row" style="height: 30vh" v-if="props.layer">
      <!-- {{ props.layer?.layer_type }} -->
      <!-- fields -->
      <q-card class="col-3">
        <q-card-section>
          <q-scroll-area style="height: 30vh">
            <q-item style="color: grey">Fields:</q-item>
            <q-item v-for="(v, k) in props.layer?.fields" :key="k" dense>
              <q-item-section>{{ k }}</q-item-section>
              <q-item-section>{{ v }}</q-item-section>
            </q-item>
          </q-scroll-area>
        </q-card-section>
      </q-card>

      <!-- dump -->
      <q-card class="col">
        <q-card-section>
          <q-item style="color: grey">Dump:</q-item>
          <q-scroll-area style="height: 30vh">
            <pre>{{ props.layer?.dump }}</pre>
          </q-scroll-area>
        </q-card-section>
      </q-card>
    </div>
    <div v-else class="row" style="height: 30vh">
      <!-- fields -->
      <q-card class="col-3">
        <q-card-section>
          <q-scroll-area style="height: 30vh">
            <q-stepper flat color="primary" animated>
            <q-step
              name="1"
              title="Select a layer"
              caption="Select a layer to view its fields (headers)"
              icon="list"
            />
          </q-stepper>
            <q-skeleton
              v-for="i in 4"
              :key="i"
              style="margin-bottom: 8px"
              animation="blink"
              :rows="10"
              :row-height="32"
              :row-width="['100%', '100%', '100%', '100%', '100%']"
            />
          </q-scroll-area>
        </q-card-section>
      </q-card>

      <!-- dump -->
      <q-card class="col">
        <q-card-section>
          <q-stepper flat color="primary" animated>
            <q-step
              name="1"
              title="Select a layer"
              caption="Select a layer to view its payload dump."
              icon="dataset"
            />
          </q-stepper>
          <q-scroll-area style="height: 30vh">
            <q-skeleton
              v-for="i in 4"
              :key="i"
              style="margin-bottom: 8px"
              animation="blink"
              :rows="10"
              :row-height="32"
              :row-width="['100%', '100%', '100%', '100%', '100%']"
            />
          </q-scroll-area>
        </q-card-section>
      </q-card>
      <q-card></q-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Layer } from 'src/model/model';
import { computed, ref } from 'vue';

interface Props {
  layer: Layer | null;
}

const props = withDefaults(defineProps<Props>(), {
  layer: () => ({} as Layer),
});
</script>
