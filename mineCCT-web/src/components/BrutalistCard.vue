<template>
  <div class="brutalist-card" ref="cardRef">
    <div class="card-header" v-if="$slots.header || title">
      <slot name="header">
        <h3 class="card-title">{{ title }}</h3>
      </slot>
    </div>
    <div class="card-content">
      <slot></slot>
    </div>
    <div class="card-decoration"></div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import gsap from 'gsap'

const props = defineProps({
  title: {
    type: String,
    default: ''
  },
  delay: {
    type: Number,
    default: 0
  }
})

const cardRef = ref(null)

onMounted(() => {
  gsap.fromTo(cardRef.value, 
    { 
      opacity: 0, 
      y: 50, 
      rotationX: -10 
    },
    { 
      opacity: 1, 
      y: 0, 
      rotationX: 0,
      duration: 0.6, 
      delay: props.delay,
      ease: "power3.out" 
    }
  )
})
</script>

<style scoped lang="scss">
.brutalist-card {
  background: var(--surface-color, #1a1a1a);
  border: 3px solid var(--border-color, #fff);
  padding: 1.5rem;
  position: relative;
  margin-bottom: 2rem;
  box-shadow: 8px 8px 0 rgba(0,0,0,0.5);
  transition: transform 0.2s;
}

.brutalist-card:hover {
  transform: translateY(-2px);
  box-shadow: 10px 10px 0 var(--primary-color, #FFD600);
}

.card-title {
  font-size: 1.2rem;
  margin-bottom: 1rem;
  border-bottom: 2px solid var(--border-color, #fff);
  padding-bottom: 0.5rem;
  display: inline-block;
  background: var(--primary-color, #FFD600);
  color: #000;
  padding: 0.2rem 0.5rem;
}

.card-decoration {
  position: absolute;
  top: 5px;
  right: 5px;
  width: 10px;
  height: 10px;
  background: var(--secondary-color, #FF5722);
  border: 1px solid #000;
}
</style>
