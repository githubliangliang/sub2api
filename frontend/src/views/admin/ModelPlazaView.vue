<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex flex-wrap items-center gap-3">
          <div class="flex-1 sm:max-w-80">
            <input
              v-model.trim="searchQuery"
              type="text"
              class="input"
              :placeholder="t('admin.modelPlaza.searchPlaceholder')"
            />
          </div>
          <Select
            v-model="platformFilter"
            :options="platformOptions"
            class="w-40"
          />
          <div class="flex flex-1 items-center justify-end">
            <button
              type="button"
              class="btn btn-secondary"
              :disabled="loading || savingPrice"
              :title="t('common.refresh')"
              @click="loadItems"
            >
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
            </button>
          </div>
        </div>
      </template>

      <template #table>
        <DataTable :columns="columns" :data="filteredItems" :loading="loading" row-key="rowKey">
          <template #cell-client_id="{ row }">
            <button
              type="button"
              class="group/copy inline-flex max-w-full items-center gap-1 rounded-md text-left font-mono text-sm font-medium text-gray-900 transition-colors hover:text-primary-600 dark:text-white dark:hover:text-primary-400"
              :title="t('admin.modelPlaza.copyClientId')"
              @click="copyId(row.client_id)"
            >
              <span class="truncate">{{ row.client_id }}</span>
              <Icon
                name="clipboard"
                size="xs"
                class="h-3.5 w-3.5 shrink-0 text-gray-400 opacity-0 transition-opacity group-hover/copy:opacity-100 group-focus-visible/copy:opacity-100 dark:text-dark-400"
              />
            </button>
          </template>

          <template #cell-upstream_ids="{ row }">
            <div class="flex flex-wrap gap-1.5">
              <button
                v-for="id in row.upstream_ids"
                :key="id"
                type="button"
                class="inline-flex max-w-full items-center rounded-md border border-gray-200 bg-gray-50 px-1.5 py-0.5 font-mono text-xs text-gray-700 transition-colors hover:border-primary-300 hover:text-primary-600 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-200 dark:hover:border-primary-700 dark:hover:text-primary-400"
                :title="t('admin.modelPlaza.copyUpstreamId')"
                @click="copyId(id)"
              >
                <span class="truncate">{{ id }}</span>
              </button>
              <span
                v-if="row.upstream_ids.length === 1 && row.upstream_ids[0] === row.client_id"
                class="self-center text-[11px] text-gray-400 dark:text-dark-500"
              >
                {{ t('admin.modelPlaza.sameAsClient') }}
              </span>
            </div>
          </template>

          <template #cell-platform="{ value }">
            <span
              :class="[
                'inline-flex items-center rounded-md border px-1.5 py-0.5 text-xs font-medium',
                platformBadgeClass(value)
              ]"
            >
              {{ platformLabel(value) }}
            </span>
          </template>

          <template #cell-pricing="{ row }">
            <div
              v-if="row.pricing"
              class="min-w-36 space-y-1 text-xs tabular-nums"
              :title="t('admin.modelPlaza.priceHint')"
            >
              <div v-if="row.pricing.manual" class="flex items-center justify-between gap-3">
                <span class="rounded bg-primary-50 px-1.5 py-0.5 text-primary-600 dark:bg-primary-900/20 dark:text-primary-400">
                  {{ t('admin.modelPlaza.manualPrice') }}
                </span>
                <button
                  type="button"
                  class="text-primary-600 hover:text-primary-700 dark:text-primary-400"
                  :title="t('admin.modelPlaza.editPrice')"
                  :disabled="loading || savingPrice"
                  @click="openPriceEditor(row)"
                >
                  {{ t('common.edit') }}
                </button>
              </div>
              <div
                v-if="row.pricing.model !== row.client_id"
                class="font-mono text-gray-400 dark:text-dark-500"
              >
                {{ row.pricing.model }}
              </div>
              <template v-for="field in priceFields" :key="field.key">
                <div
                  v-if="row.pricing[field.key] != null"
                  class="flex items-center justify-between gap-3 whitespace-nowrap"
                >
                  <span class="text-gray-500 dark:text-gray-400">
                    {{ t(`admin.modelPlaza.${field.label}`) }}
                  </span>
                  <span class="font-mono text-gray-700 dark:text-gray-200">
                    {{ formatScaled(row.pricing[field.key], 1_000_000, 2) }}
                  </span>
                </div>
              </template>
            </div>
            <button
              v-else
              type="button"
              class="btn btn-secondary px-2 py-1 text-xs"
              :title="t('admin.modelPlaza.addPrice')"
              :disabled="loading || savingPrice"
              @click="openPriceEditor(row)"
            >
              {{ t('admin.modelPlaza.addPrice') }}
            </button>
          </template>

          <template #cell-account_count="{ value }">
            <span class="tabular-nums text-sm text-gray-700 dark:text-gray-200">{{ value }}</span>
          </template>

          <template #empty>
            <div class="flex flex-col items-center py-8">
              <Icon name="inbox" size="xl" class="mb-4 h-12 w-12 text-gray-300 dark:text-dark-600" />
              <p class="text-sm font-medium text-gray-500 dark:text-gray-400">
                {{ t('admin.modelPlaza.empty') }}
              </p>
            </div>
          </template>
        </DataTable>
      </template>
    </TablePageLayout>

    <BaseDialog
      :show="editingPrice !== null"
      :title="t(editingPrice?.pricing?.manual ? 'admin.modelPlaza.editPrice' : 'admin.modelPlaza.addPrice')"
      :close-on-escape="!savingPrice"
      :show-close-button="!savingPrice"
      @close="closePriceEditor"
    >
      <form id="plaza-price-form" class="space-y-4" @submit.prevent="savePrice()">
        <div class="space-y-1">
          <p class="break-all font-mono text-sm text-gray-900 dark:text-white">{{ editingPrice?.client_id }}</p>
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ platformLabel(editingPrice?.platform ?? '') }}</p>
          <p class="text-sm font-medium text-gray-700 dark:text-gray-200">{{ t('admin.modelPlaza.price') }}</p>
          <p class="input-hint">{{ t('admin.modelPlaza.manualPriceHint') }}</p>
        </div>
        <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div v-for="field in priceFields" :key="field.key">
            <label :for="`plaza-price-${field.key}`" class="input-label">
              {{ t(`admin.modelPlaza.${field.label}`) }}
              <span v-if="field.key === 'input_price' || field.key === 'output_price'" class="text-red-500">*</span>
            </label>
            <input
              :id="`plaza-price-${field.key}`"
              v-model="priceForm[field.key]"
              type="number"
              min="0"
              step="any"
              class="input"
              :required="field.key === 'input_price' || field.key === 'output_price'"
              :disabled="savingPrice"
            />
          </div>
        </div>
      </form>
      <template #footer>
        <div class="flex flex-wrap justify-end gap-3">
          <button
            v-if="editingPrice?.pricing?.manual"
            type="button"
            class="btn btn-secondary mr-auto text-red-600 dark:text-red-400"
            data-testid="clear-reference-price"
            :disabled="savingPrice"
            @click="savePrice(true)"
          >
            {{ t('admin.modelPlaza.clearPrice') }}
          </button>
          <button type="button" class="btn btn-secondary" :disabled="savingPrice" @click="closePriceEditor">
            {{ t('common.cancel') }}
          </button>
          <button type="submit" form="plaza-price-form" class="btn btn-primary" :disabled="savingPrice">
            {{ t(savingPrice ? 'common.saving' : 'common.save') }}
          </button>
        </div>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import type { AccountModelPlazaItem } from '@/api/admin/modelPlaza'
import type { PlazaOfficialPricing } from '@/api/modelPlaza'
import BaseDialog from '@/components/common/BaseDialog.vue'
import DataTable from '@/components/common/DataTable.vue'
import Select from '@/components/common/Select.vue'
import type { Column } from '@/components/common/types'
import Icon from '@/components/icons/Icon.vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import { useClipboard } from '@/composables/useClipboard'
import { useAppStore } from '@/stores/app'
import { platformBadgeClass, platformLabel } from '@/utils/platformColors'
import { formatScaled } from '@/utils/pricing'

interface PlazaRow extends AccountModelPlazaItem {
  rowKey: string
}

const { t } = useI18n()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

const loading = ref(false)
const items = ref<PlazaRow[]>([])
const searchQuery = ref('')
const platformFilter = ref('')
const editingPrice = ref<PlazaRow | null>(null)
const savingPrice = ref(false)

const priceFields = [
  { key: 'input_price', label: 'inputPrice' },
  { key: 'output_price', label: 'outputPrice' },
  { key: 'cache_read_price', label: 'cacheReadPrice' },
  { key: 'cache_write_price', label: 'cacheWritePrice' },
  { key: 'cache_write_1h_price', label: 'cacheWrite1hPrice' }
] as const

type PriceField = typeof priceFields[number]['key']
const priceForm = ref<Record<PriceField, string | number>>({
  input_price: '',
  output_price: '',
  cache_read_price: '',
  cache_write_price: '',
  cache_write_1h_price: ''
})

const columns = computed<Column[]>(() => [
  { key: 'client_id', label: t('admin.modelPlaza.clientId'), sortable: true },
  { key: 'upstream_ids', label: t('admin.modelPlaza.upstreamIds') },
  { key: 'platform', label: t('admin.modelPlaza.platform'), sortable: true },
  { key: 'pricing', label: t('admin.modelPlaza.price') },
  { key: 'account_count', label: t('admin.modelPlaza.accountCount'), sortable: true }
])

const platformOptions = computed(() => {
  const platforms = [...new Set(items.value.map((item) => item.platform))].sort()
  return [
    { value: '', label: t('admin.modelPlaza.allPlatforms') },
    ...platforms.map((platform) => ({
      value: platform,
      label: platformLabel(platform)
    }))
  ]
})

const filteredItems = computed(() => {
  const query = searchQuery.value.toLowerCase()
  return items.value.filter((item) => {
    if (platformFilter.value && item.platform !== platformFilter.value) {
      return false
    }
    if (!query) {
      return true
    }
    return (
      item.client_id.toLowerCase().includes(query) ||
      item.upstream_ids.some((id) => id.toLowerCase().includes(query))
    )
  })
})

function copyId(id: string) {
  void copyToClipboard(id)
}

function openPriceEditor(row: PlazaRow) {
  for (const { key } of priceFields) {
    const value = row.pricing?.[key]
    priceForm.value[key] = value == null ? '' : Number((value * 1_000_000).toPrecision(15)).toString()
  }
  editingPrice.value = row
}

function closePriceEditor() {
  if (!savingPrice.value) editingPrice.value = null
}

async function savePrice(clear = false) {
  const row = editingPrice.value
  if (!row || savingPrice.value) return

  let pricing: PlazaOfficialPricing | null = null
  if (!clear) {
    const values = {} as Record<PriceField, number | null>
    for (const { key } of priceFields) {
      const raw = String(priceForm.value[key]).trim()
      if (raw === '' && key !== 'input_price' && key !== 'output_price') {
        values[key] = null
      } else if (raw === '' || !Number.isFinite(Number(raw)) || Number(raw) < 0) {
        appStore.showError(t('admin.modelPlaza.invalidPrice'))
        return
      } else {
        values[key] = Number(raw) / 1_000_000
      }
    }
    pricing = values
  }

  savingPrice.value = true
  try {
    await adminAPI.modelPlaza.updateReferencePricing({
      client_id: row.client_id,
      platform: row.platform,
      pricing
    })
    appStore.showSuccess(t('admin.modelPlaza.priceSaved'))
    editingPrice.value = null
    await loadItems()
  } catch {
    appStore.showError(t('admin.modelPlaza.savePriceFailed'))
  } finally {
    savingPrice.value = false
  }
}

async function loadItems() {
  loading.value = true
  try {
    const data = await adminAPI.modelPlaza.listAccountModels()
    items.value = (data.items ?? []).map((item) => ({
      ...item,
      rowKey: `${item.platform}:${item.client_id}`
    }))
  } catch {
    appStore.showError(t('admin.modelPlaza.loadFailed'))
    items.value = []
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void loadItems()
})
</script>
