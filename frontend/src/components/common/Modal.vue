<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="show" class="modal-overlay" @click="handleOverlayClick">
        <div class="modal-container" @click.stop>
          <!-- 標題欄 -->
          <div class="modal-header">
            <div class="flex items-center gap-3">
              <div class="w-8 h-8 bg-gradient-to-r from-blue-500 to-purple-600 rounded-full flex items-center justify-center">
                <svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                </svg>
              </div>
              <h3 class="modal-title">{{ title }}</h3>
            </div>
            <button
              @click="closeModal"
              class="modal-close"
              type="button"
              aria-label="關閉"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
              </svg>
            </button>
          </div>

          <!-- 內容區域 -->
          <div class="modal-content">
            <slot></slot>
          </div>

          <!-- 底部按鈕區域 -->
          <div v-if="$slots.footer" class="modal-footer">
            <slot name="footer"></slot>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { watch } from 'vue'

interface Props {
  show: boolean
  title?: string
  closeOnOverlay?: boolean
}

interface Emits {
  (e: 'update:show', value: boolean): void
}

const props = withDefaults(defineProps<Props>(), {
  title: '對話框',
  closeOnOverlay: true
})

const emit = defineEmits<Emits>()

const closeModal = () => {
  emit('update:show', false)
}

const handleOverlayClick = () => {
  if (props.closeOnOverlay) {
    closeModal()
  }
}

// 監聽 ESC 鍵關閉模態框
const handleEscKey = (event: KeyboardEvent) => {
  if (event.key === 'Escape' && props.show) {
    closeModal()
  }
}

// 監聽 show 屬性變化，添加/移除 ESC 鍵監聽
watch(() => props.show, (newValue) => {
  if (newValue) {
    document.addEventListener('keydown', handleEscKey)
    document.body.style.overflow = 'hidden'
  } else {
    document.removeEventListener('keydown', handleEscKey)
    document.body.style.overflow = ''
  }
}, { immediate: true })
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(135deg, rgba(0, 0, 0, 0.8), rgba(59, 130, 246, 0.3));
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.modal-container {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 1.5rem;
  box-shadow: 
    0 25px 50px -12px rgba(0, 0, 0, 0.25),
    0 0 0 1px rgba(255, 255, 255, 0.1);
  max-width: 90vw;
  max-height: 90vh;
  width: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  transform: translateY(0);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.modal-container:hover {
  transform: translateY(-2px);
  box-shadow: 
    0 32px 64px -12px rgba(0, 0, 0, 0.3),
    0 0 0 1px rgba(255, 255, 255, 0.15);
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1.5rem 2rem;
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.1), rgba(139, 92, 246, 0.1));
  border-bottom: 1px solid rgba(59, 130, 246, 0.2);
}

.modal-title {
  font-size: 1.25rem;
  font-weight: 700;
  background: linear-gradient(135deg, #3b82f6, #8b5cf6);
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
  margin: 0;
}

.modal-close {
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.2);
  color: #ef4444;
  cursor: pointer;
  padding: 0.5rem;
  border-radius: 0.75rem;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  display: flex;
  align-items: center;
  justify-content: center;
}

.modal-close:hover {
  background: rgba(239, 68, 68, 0.2);
  border-color: rgba(239, 68, 68, 0.3);
  transform: scale(1.05);
}

.modal-close:active {
  transform: scale(0.95);
}

.modal-content {
  padding: 2rem;
  overflow-y: auto;
  flex: 1;
  background: rgba(255, 255, 255, 0.8);
}

.modal-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 1rem;
  padding: 1.5rem 2rem;
  background: linear-gradient(135deg, rgba(249, 250, 251, 0.8), rgba(243, 244, 246, 0.8));
  border-top: 1px solid rgba(59, 130, 246, 0.1);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
}

/* 過渡動畫 */
.modal-enter-active,
.modal-leave-active {
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
}

.modal-enter-from {
  opacity: 0;
  transform: scale(0.8) translateY(20px);
}

.modal-leave-to {
  opacity: 0;
  transform: scale(0.8) translateY(20px);
}

/* 自定義滾動條 */
.modal-content::-webkit-scrollbar {
  width: 6px;
}

.modal-content::-webkit-scrollbar-track {
  background: rgba(59, 130, 246, 0.1);
  border-radius: 3px;
}

.modal-content::-webkit-scrollbar-thumb {
  background: linear-gradient(to bottom, #3b82f6, #8b5cf6);
  border-radius: 3px;
}

.modal-content::-webkit-scrollbar-thumb:hover {
  background: linear-gradient(to bottom, #2563eb, #7c3aed);
}

/* 響應式設計 */
@media (max-width: 640px) {
  .modal-container {
    margin: 0.5rem;
    max-width: calc(100vw - 1rem);
    max-height: calc(100vh - 1rem);
    border-radius: 1rem;
  }

  .modal-header,
  .modal-content,
  .modal-footer {
    padding: 1rem;
  }

  .modal-title {
    font-size: 1.125rem;
  }
}

/* 深色模式支援 */
@media (prefers-color-scheme: dark) {
  .modal-container {
    background: rgba(17, 24, 39, 0.95);
    border-color: rgba(59, 130, 246, 0.3);
  }

  .modal-header {
    background: linear-gradient(135deg, rgba(59, 130, 246, 0.2), rgba(139, 92, 246, 0.2));
    border-bottom-color: rgba(59, 130, 246, 0.3);
  }

  .modal-content {
    background: rgba(17, 24, 39, 0.8);
  }

  .modal-footer {
    background: linear-gradient(135deg, rgba(31, 41, 55, 0.8), rgba(55, 65, 81, 0.8));
    border-top-color: rgba(59, 130, 246, 0.2);
  }
}
</style>
