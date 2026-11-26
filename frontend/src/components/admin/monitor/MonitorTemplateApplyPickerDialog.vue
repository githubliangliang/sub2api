<template>
  <BaseDialog
    :show="show"
    :title="t('admin.channelMonitor.template.applyPickerTitle', { name: templateName })"
    @close="closeDialog"
  >
    <p class="mb-3 text-sm text-gray-600 dark:text-gray-400">
      {{ t('admin.channelMonitor.template.applyPickerHint') }}
    </p>

    <div v-if="loading" class="py-6 text-center text-sm text-gray-400">
      {{ t('common.loading') }}
    </div>

    <div v-else-if="monitors.length === 0" class="py-6 text-center text-sm text-gray-400">
      {{ t('admin.channelMonitor.template.applyPickerEmpty') }}
    </div>

    <div v-else>
      <!-- 全选/全不选 -->
      <div class="mb-2 flex items-center gap-3 text-xs">
        <button
          type="button"
          class="text-primary-600 hover:underline dark:text-primary-400"
          @click="selectAll"
        >
          {{ t('common.selectAll') }}
        </button>
        <button
          type="button"
          class="text-gray-500 hover:underline dark:text-gray-400"
          @click="selectNone"
        >
          {{ t('admin.channelMonitor.template.selectNone') }}
        </button>
        <span class="ml-auto text-gray-500 dark:text-gray-400">
          {{ t('admin.channelMonitor.template.selectedCount', {
            n: selectedIds.length,
            total: monitors.length,
          }) }}
        </span>
      </div>

      <ul class="max-h-80 divide-y divide-gray-100 overflow-y-auto rounded-lg border border-gray-200 dark:divide-dark-700 dark:border-dark-700">
        <li
          v-for="m in monitors"
          :key="m.id"
          class="flex cursor-pointer items-center gap-3 px-3 py-2 hover:bg-gray-50 dark:hover:bg-dark-800"
          @click="toggle(m.id)"
        >
          <input
            type="checkbox"
            :checked="selectedSet.has(m.id)"
            class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
            @click.stop="toggle(m.id)"
          />
          <span class="font-medium text-gray-900 dark:text-white">{{ m.name }}</span>
          <span class="text-xs text-gray-400">{{ m.provider }}</span>
          <span v-if="m.provider === 'openai'" class="text-xs text-gray-400">{{ m.api_mode }}</span>
          <span
            v-if="!m.enabled"
            class="ml-auto rounded bg-gray-100 px-1.5 py-0.5 text-xs text-gray-500 dark:bg-dark-700 dark:text-gray-400"
          >
            {{ t('admin.channelMonitor.onlyDisabled').replace(/^仅|^Only /, '') }}
          </span>
        </li>
      </ul>
    </div>

    <template #footer>
      <div class="flex justify-end gap-2">
        <button class="btn btn-secondary" @click="closeDialog">
          {{ t('common.cancel') }}
        </button>
        <button
          class="btn btn-primary"
          :disabled="loading || !loaded || submitting || selectedIds.length === 0"
          @click="handleApply"
        >
          {{ submitting
            ? t('common.submitting')
            : t('admin.channelMonitor.template.applyPickerConfirm', { n: selectedIds.length }) }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import { adminAPI } from '@/api/admin'
import type { AssociatedMonitorBrief } from '@/api/admin/channelMonitorTemplate'
import BaseDialog from '@/components/common/BaseDialog.vue'

const props = defineProps<{
  show: boolean
  templateId: number | null
  templateName: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'applied', affected: number): void
}>()

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const loaded = ref(false)
const submitting = ref(false)
const monitors = ref<AssociatedMonitorBrief[]>([])
const selectedIds = ref<number[]>([])

const selectedSet = computed(() => new Set(selectedIds.value))
let dialogVersion = 0
let listRequestId = 0

function resetDialog() {
  dialogVersion++
  listRequestId++
  loading.value = false
  loaded.value = false
  submitting.value = false
  monitors.value = []
  selectedIds.value = []
}

function isCurrentDialog(version: number, id: number) {
  return version === dialogVersion && props.show && props.templateId === id
}

function closeDialog() {
  resetDialog()
  emit('close')
}

onBeforeUnmount(resetDialog)

watch(
  () => [props.show, props.templateId] as const,
  ([show, id]) => {
    resetDialog()
    if (!show || id == null) return
    void fetchMonitors(id)
  },
  { immediate: true, flush: 'sync' },
)

async function fetchMonitors(id: number) {
  if (!props.show || props.templateId !== id) return
  const version = dialogVersion
  const requestId = ++listRequestId
  const isCurrentRequest = () => isCurrentDialog(version, id) && requestId === listRequestId
  loading.value = true
  loaded.value = false
  monitors.value = []
  selectedIds.value = []
  try {
    const { items } = await adminAPI.channelMonitorTemplate.listAssociatedMonitors(id)
    if (!isCurrentRequest()) return
    monitors.value = items
    // 默认全选
    selectedIds.value = items.map((m) => m.id)
    loaded.value = true
  } catch (err: unknown) {
    if (!isCurrentRequest()) return
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  } finally {
    if (isCurrentRequest()) loading.value = false
  }
}

function toggle(id: number) {
  const idx = selectedIds.value.indexOf(id)
  if (idx >= 0) selectedIds.value.splice(idx, 1)
  else selectedIds.value.push(id)
}

function selectAll() {
  selectedIds.value = monitors.value.map((m) => m.id)
}

function selectNone() {
  selectedIds.value = []
}

async function handleApply() {
  const id = props.templateId
  if (!props.show || id == null || !loaded.value || loading.value || selectedIds.value.length === 0 || submitting.value) return
  const version = dialogVersion
  const ids = [...selectedIds.value]
  submitting.value = true
  try {
    const { affected } = await adminAPI.channelMonitorTemplate.apply(id, ids)
    if (!isCurrentDialog(version, id)) return
    appStore.showSuccess(t('admin.channelMonitor.template.applySuccess', { n: affected }))
    emit('applied', affected)
    if (isCurrentDialog(version, id)) closeDialog()
  } catch (err: unknown) {
    if (!isCurrentDialog(version, id)) return
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  } finally {
    if (isCurrentDialog(version, id)) submitting.value = false
  }
}
</script>
