<template>
  <section class="panel-section">
    <el-row :gutter="20">
      <el-col :xs="24" :md="12" :lg="10">
        <div class="mc-panel overview-card">
          <div class="ae2-header-bar">
            <div class="header-title">能量储备</div>
            <div class="header-tag">{{ Math.floor(energyPercent) }}%</div>
          </div>

          <div class="overview-body">
            <el-progress
              :percentage="energyPercent"
              :stroke-width="14"
              :color="energyColor"
              :show-text="false"
            />

            <div class="energy-meta">
              <div class="energy-value">
                {{ formatCompact(systemStatus.energyStored) }} / {{ formatCompact(systemStatus.energyMax) }} AE
              </div>
            </div>

            <div class="stats-grid">
              <div class="energy-stat">
                <div class="label">输入</div>
                <div class="value">{{ formatRate(systemStatus.averageEnergyInput) }}</div>
              </div>
              <div class="energy-stat">
                <div class="label">消耗</div>
                <div class="value">{{ formatRate(systemStatus.energyUsage) }}</div>
              </div>
              <div class="energy-stat">
                <div class="label">变化</div>
                <div class="value" :class="netRateClass">{{ formatRate(systemStatus.netEnergyRate, true) }}</div>
              </div>
            </div>
          </div>
        </div>
      </el-col>

      <el-col :xs="24" :md="12" :lg="14">
        <div class="mc-panel overview-card">
          <div class="ae2-header-bar">
            <div class="header-title">存储总览</div>
            <div class="header-tag">{{ Math.floor(storagePercent) }}%</div>
          </div>

          <div class="overview-body">
            <div class="storage-split mc-slot" role="img">
              <div class="storage-segment internal" :style="{ width: `${storageInternalRatio}%` }">
                <div class="storage-fill" :style="{ width: `${storageInternalUsage}%`, background: '#3dd6a5' }"></div>
              </div>
              <div class="storage-segment external" :style="{ width: `${storageExternalRatio}%` }">
                <div class="storage-fill" :style="{ width: `${storageExternalUsage}%`, background: '#5d8aff' }"></div>
              </div>
            </div>

            <div class="storage-legend">
              <div class="legend-item"><span class="swatch" style="background:#3dd6a5"></span> 内部</div>
              <div class="legend-item"><span class="swatch" style="background:#5d8aff"></span> 外部</div>
            </div>

            <div class="storage-stats-grid">
              <div class="mc-slot storage-block">
                <div class="block-title">物品存储</div>
                <div class="block-row">{{ formatCompact(systemStatus.storage.itemUsed) }} / {{ formatCompact(systemStatus.storage.itemTotal) }}</div>
              </div>
              <div class="mc-slot storage-block">
                <div class="block-title">流体存储</div>
                <div class="block-row">{{ formatCompact(systemStatus.storage.fluidUsed) }} / {{ formatCompact(systemStatus.storage.fluidTotal) }}</div>
              </div>
            </div>
          </div>
        </div>
      </el-col>
    </el-row>
  </section>
</template>

<script setup>
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

<style scoped>
.overview-card {
  margin-bottom: 20px;
}

.overview-body {
  padding: 15px;
}

.header-tag {
  background: #aaaaaa;
  border: 2px solid var(--ae2-border-dark);
  box-shadow: inset 1px 1px 0px #ffffff;
  padding: 2px 8px;
  font-size: 0.9rem;
  color: #333;
  font-weight: bold;
}

.energy-meta {
  margin: 10px 0;
  text-align: center;
}

.energy-value {
  font-size: 1rem;
  font-weight: bold;
  color: var(--ae2-text);
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
  margin-top: 15px;
}

.energy-stat {
  flex-direction: column;
  padding: 8px !important;
}

.energy-stat .label {
  color: #bbbbbb !important;
  font-size: 0.75rem;
  margin-bottom: 4px;
}

.energy-stat .value {
  font-size: 0.9rem;
  font-weight: bold;
}

.storage-split {
  display: flex;
  overflow: hidden;
  margin-bottom: 10px;
  padding: 0 !important;
}

.storage-segment {
  height: 100%;
  position: relative;
}

.storage-fill {
  height: 100%;
}

.storage-legend {
  display: flex;
  gap: 15px;
  justify-content: center;
  margin-bottom: 15px;
  font-size: 0.85rem;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 6px;
}

.swatch {
  width: 10px;
  height: 10px;
  border: 1px solid var(--ae2-border-dark);
}

.storage-stats-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}

.storage-block {
  flex-direction: column;
  padding: 10px !important;
  text-align: center;
}

.block-title {
  font-size: 0.8rem;
  color: #ccc;
  margin-bottom: 4px;
  text-shadow: 1px 1px 0px #373737;
}

.block-row {
  font-size: 0.95rem;
  font-weight: bold;
  color: white;
  text-shadow: 1px 1px 0px #373737;
}
</style>
