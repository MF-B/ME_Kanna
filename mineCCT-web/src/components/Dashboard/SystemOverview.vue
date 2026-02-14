<template>
  <section class="panel-section">
    <div class="panel-title">AE2 监控面板</div>

    <el-row :gutter="20">
      <el-col :xs="24" :md="12" :lg="10">
        <el-card class="energy-card" shadow="hover">
          <div class="energy-header">
            <div class="energy-title">能量储备</div>
            <el-tag size="small" effect="dark" :type="energyPercent < 20 ? 'danger' : 'success'">
              {{ Math.floor(energyPercent) }}%
            </el-tag>
          </div>

          <el-progress
            :percentage="energyPercent"
            :stroke-width="10"
            :color="energyColor"
            :show-text="false"
          />

          <div class="energy-meta">
            <div class="energy-value">
              {{ formatCompact(systemStatus.energyStored) }} / {{ formatCompact(systemStatus.energyMax) }} AE
            </div>
            <div class="energy-updated" v-if="systemStatus.lastUpdated">
              更新: {{ formatTime(systemStatus.lastUpdated) }}
            </div>
          </div>

          <div class="energy-grid">
            <div class="energy-stat">
              <div class="label">输入均值</div>
              <div class="value">{{ formatRate(systemStatus.averageEnergyInput) }}</div>
            </div>
            <div class="energy-stat">
              <div class="label">消耗速率</div>
              <div class="value">{{ formatRate(systemStatus.energyUsage) }}</div>
            </div>
            <div class="energy-stat">
              <div class="label">净变化</div>
              <div class="value" :class="netRateClass">{{ formatRate(systemStatus.netEnergyRate, true) }}</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :xs="24" :md="12" :lg="14">
        <el-card class="storage-card" shadow="hover">
          <div class="energy-header">
            <div class="energy-title">库存总览</div>
            <el-tag size="small" effect="dark" :type="storagePercent < 80 ? 'success' : 'warning'">
              {{ Math.floor(storagePercent) }}%
            </el-tag>
          </div>

          <div class="storage-split" role="img" aria-label="内部与外部存储容量占比">
            <div class="storage-segment internal" :style="{ width: `${storageInternalRatio}%` }">
              <div class="storage-fill" :style="{ width: `${storageInternalUsage}%` }"></div>
            </div>
            <div class="storage-segment external" :style="{ width: `${storageExternalRatio}%` }">
              <div class="storage-fill" :style="{ width: `${storageExternalUsage}%` }"></div>
            </div>
          </div>

          <div class="storage-legend">
            <div class="legend-item">
              <span class="legend-swatch internal"></span>
              内部 {{ formatCompact(systemStatus.storage.itemTotal) }}
            </div>
            <div class="legend-item">
              <span class="legend-swatch external"></span>
              外部 {{ formatCompact(systemStatus.storage.itemExternalTotal) }}
            </div>
          </div>

          <div class="energy-meta">
            <div class="energy-value">
              {{ formatCompact(storageTotalUsed) }} / {{ formatCompact(storageTotalCapacity) }} 物品存储
            </div>
            <div class="energy-updated" v-if="systemStatus.lastUpdated">
              更新: {{ formatTime(systemStatus.lastUpdated) }}
            </div>
          </div>

          <div class="storage-grid">
            <div class="storage-block">
              <div class="block-title">物品存储</div>
              <div class="block-row">已用 {{ formatCompact(systemStatus.storage.itemUsed) }} / 总计 {{ formatCompact(systemStatus.storage.itemTotal) }}</div>
              <div class="block-row muted">外部 {{ formatCompact(systemStatus.storage.itemExternalUsed) }} / {{ formatCompact(systemStatus.storage.itemExternalTotal) }}</div>
            </div>
            <div class="storage-block">
              <div class="block-title">流体存储</div>
              <div class="block-row">已用 {{ formatCompact(systemStatus.storage.fluidUsed) }} / 总计 {{ formatCompact(systemStatus.storage.fluidTotal) }}</div>
              <div class="block-row muted">外部 {{ formatCompact(systemStatus.storage.fluidExternalUsed) }} / {{ formatCompact(systemStatus.storage.fluidExternalTotal) }}</div>
            </div>
          </div>
        </el-card>
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
