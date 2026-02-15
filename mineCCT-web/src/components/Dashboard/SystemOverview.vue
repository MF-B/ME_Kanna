<template>
  <div class="system-overview">
    <el-row :gutter="40">
      <el-col :xs="24" :md="12">
        <BrutalistCard :title="t('SYSTEM.ENERGY')" :delay="0.2" class="h-100">
          <div class="d-flex flex-column h-100 justify-content-between">
            <div>
              <div class="big-stat">
                <span class="value">{{ Math.floor(energyPercent) }}</span>
                <span class="unit">%</span>
              </div>
              
              <div class="brutalist-progress">
                <div class="bar" :style="{ width: energyPercent + '%', background: energyColor }"></div>
              </div>
              
              <div class="stat-row single-line">
                <span class="val">{{ formatCompact(systemStatus.energyStored) }} / {{ formatCompact(systemStatus.energyMax) }} AE</span>
              </div>
            </div>

            <div class="grid-stats">
              <div class="g-stat">
                <span class="lbl">{{ t('SYSTEM.IN') }}</span>
                <span class="num">{{ formatRate(systemStatus.averageEnergyInput) }}</span>
              </div>
              <div class="g-stat">
                <span class="lbl">{{ t('SYSTEM.OUT') }}</span>
                <span class="num">{{ formatRate(systemStatus.energyUsage) }}</span>
              </div>
            </div>
          </div>
        </BrutalistCard>
      </el-col>

      <el-col :xs="24" :md="12">
        <BrutalistCard :title="t('SYSTEM.STORAGE')" :delay="0.4" class="h-100">
          <div class="d-flex flex-column h-100 justify-content-between">
            <div>
              <div class="big-stat">
                <span class="value">{{ totalStoragePercent }}</span>
                <span class="unit">%</span>
              </div>
              
              <div class="brutalist-progress">
                 <div class="bar" :style="{ width: totalStoragePercent + '%', background: 'var(--accent-color)' }"></div>
              </div>

              <div class="stat-row single-line">
                 <span class="val">{{ formatCompact(totalStorageUsed) }} / {{ formatCompact(totalStorageCapacity) }}</span>
              </div>
            </div>

            <div class="storage-details">
               <div class="detail-box">
                  <h4>{{ t('SYSTEM.INTERNAL') }}</h4>
                  <p>{{ formatCompact(systemStatus.storage.itemUsed) }} / {{ formatCompact(systemStatus.storage.itemTotal) }}</p>
               </div>
               <div class="detail-box">
                  <h4>{{ t('SYSTEM.EXTERNAL') }}</h4>
                  <p>{{ formatCompact(systemStatus.storage.fluidUsed) }} / {{ formatCompact(systemStatus.storage.fluidTotal) }}</p>
               </div>
            </div>
          </div>
        </BrutalistCard>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import BrutalistCard from '../BrutalistCard.vue'
import { useI18n } from '../../composables/useI18n'

const { t } = useI18n()

const props = defineProps({
  systemStatus: { type: Object, required: true },
  energyPercent: { type: Number, required: true },
  energyColor: { type: String, required: true },
  storagePercent: { type: Number, required: true },
  storageTotalUsed: { type: Number, required: true },
  storageTotalCapacity: { type: Number, required: true },
  storageInternalRatio: { type: Number, required: true },
  storageExternalRatio: { type: Number, required: true },
  storageInternalUsage: { type: Number, required: true },
  storageExternalUsage: { type: Number, required: true },
  netRateClass: { type: String, required: true },
  formatCompact: { type: Function, required: true },
  formatRate: { type: Function, required: true },
  formatTime: { type: Function, required: true }
})

// Calculate total storage percent from existing props
const totalStorageUsed = computed(() => props.storageTotalUsed)
const totalStorageCapacity = computed(() => props.storageTotalCapacity)
const totalStoragePercent = computed(() => {
  if (!totalStorageCapacity.value) return 0
  return Math.floor((totalStorageUsed.value / totalStorageCapacity.value) * 100)
})
</script>

<style scoped lang="scss">
.big-stat {
  font-size: 4rem;
  font-weight: 900;
  line-height: 1;
  margin-bottom: 1rem;
  color: var(--text-color);
  
  .unit {
    font-size: 1.5rem;
    color: #888;
    margin-left: 0.5rem;
  }
}

.brutalist-progress {
  height: 24px;
  background: var(--bg-color);
  border: 2px solid var(--border-color);
  margin-bottom: 1rem;
  position: relative;
  
  .bar {
    height: 100%;
    transition: width 0.5s cubic-bezier(0.19, 1, 0.22, 1);
  }
}

.stat-row {
  display: flex;
  justify-content: space-between;
  margin-bottom: 1.5rem;
  font-family: monospace;
  color: var(--text-color);
  
  &.single-line {
    justify-content: center; /* Center the single line text */
    margin-bottom: 2rem;
  }
  
  .label { color: #888; display: block; font-size: 0.8rem; }
  .val { font-weight: bold; font-size: 1.1rem; }
  .right { text-align: right; }
}

.grid-stats {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
  
  .g-stat {
    border: 2px solid var(--border-color);
    padding: 10px;
    text-align: center;
    background: var(--bg-color);
    color: var(--text-color);
    
    .lbl { display: block; font-size: 0.7rem; color: #888; margin-bottom: 5px; }
    .num { display: block; font-weight: bold; font-size: 1rem; }
  }
}

.multi-progress {
  display: flex;
  height: 24px;
  border: 2px solid var(--border-color);
  margin-bottom: 1rem;
  background: var(--bg-color);
  
  .seg { height: 100%; position: relative; }
  .fill { height: 100%; transition: width 0.5s; }
  
  .internal .fill { background: var(--accent-color); }
  .external .fill { background: var(--secondary-color); }
}

.legend {
  display: flex;
  gap: 20px;
  margin-bottom: 1.5rem;
  font-size: 0.8rem;
  font-weight: bold;
  color: var(--text-color);
  
  .dot {
    width: 12px;
    height: 12px;
    display: inline-block;
    margin-right: 5px;
    border: 1px solid var(--border-color);
  }
  .int { background: var(--accent-color); }
  .ext { background: var(--secondary-color); }
}

.storage-details {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  
  .detail-box {
    border: 2px solid var(--border-color);
    padding: 10px;
    background: var(--bg-color);
    color: var(--text-color);
    
    h4 { margin: 0 0 5px 0; font-size: 0.8rem; color: #888; }
    p { margin: 0; font-weight: bold; }
  }
}
</style>
