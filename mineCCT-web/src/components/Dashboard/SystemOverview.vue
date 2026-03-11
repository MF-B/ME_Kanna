<template>
  <div class="system-overview">
    <el-row :gutter="40">
      <el-col :xs="24" :md="12">
        <BrutalistCard title="ENERGY.CORE" :delay="0.2">
          <div class="big-stat">
            <span class="value">{{ Math.floor(energyPercent) }}</span>
            <span class="unit">%</span>
          </div>
          
          <div class="brutalist-progress">
            <div class="bar" :style="{ width: energyPercent + '%', background: energyColor }"></div>
          </div>
          
          <div class="stat-row">
            <div class="stat-item">
              <span class="label">STORED</span>
              <span class="val">{{ formatCompact(systemStatus.energyStored) }} AE</span>
            </div>
            <div class="stat-item right">
              <span class="label">MAX</span>
              <span class="val">{{ formatCompact(systemStatus.energyMax) }} AE</span>
            </div>
          </div>

          <div class="grid-stats">
            <div class="g-stat">
              <span class="lbl">IN</span>
              <span class="num">{{ formatRate(systemStatus.averageEnergyInput) }}</span>
            </div>
            <div class="g-stat">
              <span class="lbl">OUT</span>
              <span class="num">{{ formatRate(systemStatus.energyUsage) }}</span>
            </div>
            <div class="g-stat">
              <span class="lbl">NET</span>
              <span class="num" :class="netRateClass">{{ formatRate(systemStatus.netEnergyRate, true) }}</span>
            </div>
          </div>
        </BrutalistCard>
      </el-col>

      <el-col :xs="24" :md="12">
        <BrutalistCard title="STORAGE.MATRIX" :delay="0.4">
          <div class="big-stat">
            <span class="value">{{ Math.floor(storagePercent) }}</span>
            <span class="unit">%</span>
          </div>
          
          <div class="multi-progress">
             <div class="seg internal" :style="{ width: storageInternalRatio + '%' }">
               <div class="fill" :style="{ width: storageInternalUsage + '%' }"></div>
             </div>
             <div class="seg external" :style="{ width: storageExternalRatio + '%' }">
               <div class="fill" :style="{ width: storageExternalUsage + '%' }"></div>
             </div>
          </div>

          <div class="legend">
             <span class="dot int"></span> INTERNAL
             <span class="dot ext"></span> EXTERNAL
          </div>

          <div class="storage-details">
             <div class="detail-box">
                <h4>ITEM</h4>
                <p>{{ formatCompact(systemStatus.storage.itemUsed) }} / {{ formatCompact(systemStatus.storage.itemTotal) }}</p>
             </div>
             <div class="detail-box">
                <h4>FLUID</h4>
                <p>{{ formatCompact(systemStatus.storage.fluidUsed) }} / {{ formatCompact(systemStatus.storage.fluidTotal) }}</p>
             </div>
          </div>
        </BrutalistCard>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import BrutalistCard from '../BrutalistCard.vue'

defineProps({
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
</script>

<style scoped lang="scss">
.big-stat {
  font-size: 4rem;
  font-weight: 900;
  line-height: 1;
  margin-bottom: 1rem;
  
  .unit {
    font-size: 1.5rem;
    color: #666;
    margin-left: 0.5rem;
  }
}

.brutalist-progress {
  height: 24px;
  background: #333;
  border: 2px solid #fff;
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
  
  .label { color: #888; display: block; font-size: 0.8rem; }
  .val { font-weight: bold; font-size: 1.1rem; }
  .right { text-align: right; }
}

.grid-stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
  
  .g-stat {
    border: 2px solid #333;
    padding: 10px;
    text-align: center;
    background: #000;
    
    .lbl { display: block; font-size: 0.7rem; color: #666; margin-bottom: 5px; }
    .num { display: block; font-weight: bold; font-size: 1rem; }
  }
}

.multi-progress {
  display: flex;
  height: 24px;
  border: 2px solid #fff;
  margin-bottom: 1rem;
  background: #333;
  
  .seg { height: 100%; position: relative; }
  .fill { height: 100%; transition: width 0.5s; }
  
  .internal .fill { background: var(--accent-color, #00E676); }
  .external .fill { background: var(--secondary-color, #FF5722); }
}

.legend {
  display: flex;
  gap: 20px;
  margin-bottom: 1.5rem;
  font-size: 0.8rem;
  font-weight: bold;
  
  .dot {
    width: 12px;
    height: 12px;
    display: inline-block;
    margin-right: 5px;
    border: 1px solid #fff;
  }
  .int { background: var(--accent-color); }
  .ext { background: var(--secondary-color); }
}

.storage-details {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  
  .detail-box {
    border: 2px solid #fff;
    padding: 10px;
    background: #000;
    
    h4 { margin: 0 0 5px 0; font-size: 0.8rem; color: #888; }
    p { margin: 0; font-weight: bold; }
  }
}
</style>
