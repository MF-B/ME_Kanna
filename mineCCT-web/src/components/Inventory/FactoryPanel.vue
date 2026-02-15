<template>
  <div class="factory-panel-container">
    <div class="factory-list">
      <div
        v-for="factory in factories"
        :key="factory.id"
        class="factory-row"
      >
        <FactoryCard
          :factory="factory"
          @command="(payload) => $emit('command', payload)"
        />
      </div>
    </div>

    <div v-if="factories.length === 0 && connected" class="empty-msg">
      WAITING FOR DATA LINK...
    </div>

    <div v-if="!connected" class="empty-msg offline">
      SYSTEM OFFLINE - RETRYING...
    </div>
  </div>
</template>

<script setup>
import FactoryCard from '../FactoryCard.vue'

defineProps({
  connected: { type: Boolean, required: true },
  factories: { type: Array, required: true }
})

defineEmits(['command'])
</script>

<style scoped>
.factory-list {
  display: flex;
  flex-direction: column;
  gap: 15px;
  padding: 10px;
}

.factory-row {
  width: 100%;
}

.empty-msg {
  text-align: center;
  padding: 50px;
  color: #666;
  font-weight: bold;
  font-family: monospace;
}

.offline {
  color: var(--secondary-color, #FF5722);
}
</style>
