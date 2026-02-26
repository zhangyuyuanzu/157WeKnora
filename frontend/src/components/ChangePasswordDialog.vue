<template>
  <t-dialog
    v-model:visible="dialogVisible"
    :header="t('auth.changePassword')"
    width="480px"
    :confirm-btn="{ content: t('common.confirm'), loading: submitting, theme: 'primary' }"
    :cancel-btn="t('common.cancel')"
    :on-confirm="handleSubmit"
    :on-cancel="handleClose"
    @closed="handleClosed"
  >
    <t-form
      ref="formRef"
      :data="formData"
      :rules="formRules"
      label-align="top"
      @submit="onSubmit"
    >
      <t-form-item :label="t('auth.oldPassword')" name="oldPassword">
        <t-input
          v-model="formData.oldPassword"
          type="password"
          :placeholder="t('auth.oldPasswordPlaceholder')"
        />
      </t-form-item>
      
      <t-form-item :label="t('auth.newPassword')" name="newPassword">
        <t-input
          v-model="formData.newPassword"
          type="password"
          :placeholder="t('auth.newPasswordPlaceholder')"
        />
      </t-form-item>
      
      <t-form-item :label="t('auth.confirmPassword')" name="confirmPassword">
        <t-input
          v-model="formData.confirmPassword"
          type="password"
          :placeholder="t('auth.confirmPasswordPlaceholder')"
        />
      </t-form-item>
    </t-form>
  </t-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { MessagePlugin, FormRule, SubmitContext } from 'tdesign-vue-next'
import { changePassword } from '@/api/auth'

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:visible', 'success'])

const { t } = useI18n()
const formRef = ref()
const submitting = ref(false)

const dialogVisible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val)
})

const formData = ref({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

// 表单验证规则
const formRules = computed((): Record<string, FormRule[]> => {
  return {
    oldPassword: [
      { required: true, message: t('auth.oldPasswordRequired'), type: 'error' }
    ],
    newPassword: [
      { required: true, message: t('auth.passwordRequired'), type: 'error' },
      { min: 8, message: t('auth.passwordMinLength'), type: 'error' },
      { max: 32, message: t('auth.passwordMaxLength'), type: 'error' }
    ],
    confirmPassword: [
      { required: true, message: t('auth.confirmPasswordRequired'), type: 'error' },
      {
        validator: (val) => val === formData.value.newPassword,
        message: t('auth.passwordMismatch'),
        type: 'error'
      }
    ]
  }
})

const handleClose = () => {
  dialogVisible.value = false
}

const handleClosed = () => {
  formData.value = {
    oldPassword: '',
    newPassword: '',
    confirmPassword: ''
  }
  formRef.value?.reset()
}

const handleSubmit = () => {
  formRef.value?.submit()
}

const onSubmit = async ({ validateResult }: SubmitContext) => {
  if (validateResult === true) {
    submitting.value = true
    try {
      const res = await changePassword(formData.value.oldPassword, formData.value.newPassword)
      if (res.success) {
        MessagePlugin.success(t('auth.changePasswordSuccess'))
        emit('success')
        dialogVisible.value = false
      } else {
        MessagePlugin.error(res.message || t('common.operationFailed'))
      }
    } catch (error: any) {
      MessagePlugin.error(error.message || t('common.operationFailed'))
    } finally {
      submitting.value = false
    }
  }
}
</script>
