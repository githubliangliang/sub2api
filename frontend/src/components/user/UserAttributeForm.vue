<template>
  <div v-if="attributes.length > 0" class="space-y-4">
    <div v-for="attr in attributes" :key="attr.id">
      <label class="input-label">
        {{ attr.name }}
        <span v-if="attr.required" class="text-red-500">*</span>
      </label>

      <!-- Text Input -->
      <input
        v-if="attr.type === 'text' || attr.type === 'email' || attr.type === 'url'"
        v-model="localValues[attr.id]"
        :type="attr.type === 'text' ? 'text' : attr.type"
        :required="attr.required"
        :disabled="!valuesReady"
        :placeholder="attr.placeholder"
        class="input"
        @input="emitChange"
      />

      <!-- Number Input -->
      <input
        v-else-if="attr.type === 'number'"
        :value="localValues[attr.id]"
        type="number"
        :required="attr.required"
        :disabled="!valuesReady"
        :placeholder="attr.placeholder"
        :min="attr.validation?.min"
        :max="attr.validation?.max"
        class="input"
        @input="localValues[attr.id] = ($event.target as HTMLInputElement).value; emitChange()"
      />

      <!-- Date Input -->
      <input
        v-else-if="attr.type === 'date'"
        v-model="localValues[attr.id]"
        type="date"
        :required="attr.required"
        :disabled="!valuesReady"
        class="input"
        @input="emitChange"
      />

      <!-- Textarea -->
      <textarea
        v-else-if="attr.type === 'textarea'"
        v-model="localValues[attr.id]"
        :required="attr.required"
        :disabled="!valuesReady"
        :placeholder="attr.placeholder"
        rows="3"
        class="input"
        @input="emitChange"
      />

      <!-- Select -->
      <Select
        v-else-if="attr.type === 'select'"
        v-model="localValues[attr.id]"
        :options="attr.options || []"
        :disabled="!valuesReady"
        @change="emitChange"
      />

      <!-- Multi-Select (Checkboxes) -->
      <div v-else-if="attr.type === 'multi_select'" class="space-y-2">
        <label
          v-for="opt in attr.options"
          :key="opt.value"
          class="flex items-center gap-2"
        >
          <input
            type="checkbox"
            :value="opt.value"
            :checked="isOptionSelected(attr.id, opt.value)"
            :disabled="!valuesReady"
            @change="toggleMultiSelectOption(attr.id, opt.value)"
            class="h-4 w-4 rounded border-gray-300 text-primary-600"
          />
          <span class="text-sm text-gray-700 dark:text-gray-300">{{ opt.label }}</span>
        </label>
      </div>

      <!-- Description -->
      <p v-if="attr.description" class="input-hint">{{ attr.description }}</p>
    </div>
  </div>

  <!-- Loading State -->
  <div v-else-if="loading" class="flex justify-center py-4">
    <svg class="h-5 w-5 animate-spin text-gray-400" fill="none" viewBox="0 0 24 24">
      <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
      <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
    </svg>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onBeforeUnmount } from 'vue'
import { adminAPI } from '@/api/admin'
import type { UserAttributeDefinition, UserAttributeValuesMap } from '@/types'
import Select from '@/components/common/Select.vue'

interface Props {
  userId?: number
  modelValue: UserAttributeValuesMap
}

interface Emits {
  (e: 'update:modelValue', value: UserAttributeValuesMap): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const loading = ref(false)
const attributes = ref<UserAttributeDefinition[]>([])
const localValues = ref<UserAttributeValuesMap>({})
const valuesReady = ref(!props.userId)
let requestVersion = 0
let active = true

const loadAttributes = async () => {
  loading.value = true
  try {
    const definitions = await adminAPI.userAttributes.listEnabledDefinitions()
    if (active) attributes.value = definitions
  } catch (error) {
    if (active) console.error('Failed to load attributes:', error)
  } finally {
    if (active) loading.value = false
  }
}

const loadUserValues = async (userId: number, version: number) => {
  const isCurrent = () => active && version === requestVersion && props.userId === userId
  try {
    const values = await adminAPI.userAttributes.getUserAttributeValues(userId)
    if (!isCurrent()) return
    const valuesMap: UserAttributeValuesMap = {}
    values.forEach(v => {
      valuesMap[v.attribute_id] = v.value
    })
    localValues.value = { ...valuesMap }
    valuesReady.value = true
    emit('update:modelValue', { ...valuesMap })
  } catch (error) {
    if (isCurrent()) console.error('Failed to load user attribute values:', error)
  }
}

const emitChange = () => {
  if (!valuesReady.value) return
  emit('update:modelValue', { ...localValues.value })
}

const isOptionSelected = (attrId: number, optionValue: string): boolean => {
  const value = localValues.value[attrId]
  if (!value) return false
  try {
    const arr = JSON.parse(value)
    return Array.isArray(arr) && arr.includes(optionValue)
  } catch {
    return false
  }
}

const toggleMultiSelectOption = (attrId: number, optionValue: string) => {
  let arr: string[] = []
  const value = localValues.value[attrId]
  if (value) {
    try {
      arr = JSON.parse(value)
      if (!Array.isArray(arr)) arr = []
    } catch {
      arr = []
    }
  }

  const index = arr.indexOf(optionValue)
  if (index > -1) {
    arr.splice(index, 1)
  } else {
    arr.push(optionValue)
  }

  localValues.value[attrId] = JSON.stringify(arr)
  emitChange()
}

watch(() => props.modelValue, (newVal) => {
  if (valuesReady.value) {
    localValues.value = { ...newVal }
  }
}, { immediate: true })

watch(() => props.userId, (newUserId, oldUserId, onCleanup) => {
  const version = ++requestVersion
  onCleanup(() => { requestVersion++ })
  valuesReady.value = !newUserId
  localValues.value = !newUserId && !oldUserId ? { ...props.modelValue } : {}
  if (newUserId) {
    loadUserValues(newUserId, version)
  }
}, { immediate: true, flush: 'sync' })

onBeforeUnmount(() => {
  active = false
  requestVersion++
})

onMounted(() => {
  loadAttributes()
})
</script>
