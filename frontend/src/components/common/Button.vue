<template>
  <button
    :class="['btn', `btn-${type}`, { 'btn-disabled': disabled }]"
    :disabled="disabled"
    @click="handleClick"
  >
    <slot>{{ text }}</slot>
  </button>
</template>

<script setup lang="ts">
interface Props {
  text?: string
  type?: 'primary' | 'secondary' | 'danger'
  disabled?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  text: 'Button',
  type: 'primary',
  disabled: false
})

const emit = defineEmits<{
  click: [event: MouseEvent]
}>()

const handleClick = (event: MouseEvent) => {
  if (!props.disabled) {
    emit('click', event)
  }
}
</script>

<style scoped>
.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.2s;
}

.btn-primary {
  background-color: #409eff;
  color: white;
}

.btn-primary:hover:not(.btn-disabled) {
  background-color: #337ecc;
}

.btn-secondary {
  background-color: #909399;
  color: white;
}

.btn-secondary:hover:not(.btn-disabled) {
  background-color: #73767a;
}

.btn-danger {
  background-color: #f56c6c;
  color: white;
}

.btn-danger:hover:not(.btn-disabled) {
  background-color: #dd6161;
}

.btn-disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
