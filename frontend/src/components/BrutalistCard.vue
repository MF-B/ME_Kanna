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
      ease: "power3.out",
      onComplete: () => {
        gsap.set(cardRef.value, { clearProps: 'transform' })
      } 
    }
  )
})
</script>

<style scoped lang="scss">
.brutalist-card {
  background: var(--surface-color);
  border: 3px solid var(--border-color);
  padding: 1.5rem;
  position: relative;
  margin-bottom: 2rem;
  box-shadow: 4px 4px 0 #888; /* Match sidebar button shadow size */
  transition: all 0.1s;
  color: var(--text-color);
}

.brutalist-card:hover {
  transform: translate(2px, 2px);
  box-shadow: 2px 2px 0 #888;
}

.card-title {
  font-size: 1.2rem;
  margin-bottom: 1rem;
  border-bottom: 2px solid var(--border-color);
  padding-bottom: 0.5rem;
  display: inline-block;
  background: var(--primary-color);
  color: #000;
  padding: 0.2rem 0.5rem;
}


</style>
