<template>
  <section class="panel-section">
    <div class="panel-title">工厂产能与库存</div>

    <el-row :gutter="20">
      <el-col
        v-for="factory in factories"
        :key="factory.id"
        :xs="24" :sm="12" :md="8" :lg="6"
        style="margin-bottom: 20px;"
      >
        <FactoryCard
          :factory="factory"
          @command="(payload) => $emit('command', payload)"
        />
      </el-col>
    </el-row>

    <el-empty
      v-if="factories.length === 0 && connected"
      description="等待 AE 网络数据上报..."
    />

    <el-empty
      v-if="!connected"
      description="与 Go 后端断开连接，正在重试..."
      :image-size="100"
    />
  </section>
</template>

<script setup>
import FactoryCard from '../FactoryCard.vue'

defineProps({
  connected: { type: Boolean, required: true },
  factories: { type: Array, required: true }
})

defineEmits(['command'])
</script>
