<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import { sendKBUpdateNotification } from '@/api/knowledge-base/index';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

const props = defineProps<{
  visible: boolean;
  kbId: string;
  kbName: string;
}>();

const emit = defineEmits<{
  (e: 'update:visible', val: boolean): void;
}>();

const dialogVisible = computed({
  get: () => props.visible,
  set: (val: boolean) => emit('update:visible', val),
});

// 表单字段
const recipientInput = ref('');
const recipients = ref<string[]>([]);
const message = ref('');
const updateSummary = ref('');
const sending = ref(false);

// 重置表单
const resetForm = () => {
  recipientInput.value = '';
  recipients.value = [];
  message.value = '';
  updateSummary.value = '';
};

watch(() => props.visible, (val) => {
  if (val) {
    resetForm();
  }
});

// 验证邮箱格式
const isValidEmail = (email: string): boolean => {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim());
};

// 添加收件人（回车或点击添加）
const addRecipient = () => {
  const raw = recipientInput.value.trim();
  if (!raw) return;
  // 支持用逗号、分号、空格分隔多个邮箱
  const emails = raw.split(/[,;，；\s]+/).filter(Boolean);
  let addedCount = 0;
  for (const email of emails) {
    const trimmed = email.trim();
    if (!trimmed) continue;
    if (!isValidEmail(trimmed)) {
      MessagePlugin.warning(t('emailNotification.invalidEmail', { email: trimmed }));
      continue;
    }
    if (recipients.value.includes(trimmed)) {
      MessagePlugin.info(t('emailNotification.duplicateEmail'));
      continue;
    }
    recipients.value.push(trimmed);
    addedCount++;
  }
  if (addedCount > 0) {
    recipientInput.value = '';
  }
};

// 移除收件人
const removeRecipient = (index: number) => {
  recipients.value.splice(index, 1);
};

// 表单是否可以提交
const canSubmit = computed(() => {
  return recipients.value.length > 0 && message.value.trim().length > 0;
});

// 发送通知
const handleConfirm = async () => {
  if (!canSubmit.value) {
    if (recipients.value.length === 0) {
      MessagePlugin.warning(t('emailNotification.recipientRequired'));
    } else {
      MessagePlugin.warning(t('emailNotification.messageRequired'));
    }
    return;
  }

  sending.value = true;
  try {
    const res: any = await sendKBUpdateNotification({
      knowledge_base_id: props.kbId,
      recipients: recipients.value,
      message: message.value.trim(),
      update_summary: updateSummary.value.trim() || undefined,
    });

    const data = res?.data;
    if (data) {
      if (data.fail_count === 0) {
        MessagePlugin.success(t('emailNotification.sendSuccess', { count: data.success_count }));
      } else if (data.success_count > 0) {
        MessagePlugin.warning(
          t('emailNotification.sendPartial', {
            success: data.success_count,
            fail: data.fail_count,
          })
        );
      } else {
        MessagePlugin.error(t('emailNotification.sendAllFailed'));
      }
    } else {
      MessagePlugin.success(t('emailNotification.sendSuccess', { count: recipients.value.length }));
    }
    dialogVisible.value = false;
  } catch (error: any) {
    const errorMsg = error?.message || error?.error?.message || t('emailNotification.sendFailed');
    MessagePlugin.error(errorMsg);
  } finally {
    sending.value = false;
  }
};

const handleCancel = () => {
  dialogVisible.value = false;
};
</script>

<template>
  <t-dialog
    v-model:visible="dialogVisible"
    :header="t('emailNotification.title')"
    :confirm-btn="{
      content: t('emailNotification.send'),
      theme: 'primary',
      loading: sending,
      disabled: !canSubmit,
    }"
    :cancel-btn="{ content: t('common.cancel') }"
    @confirm="handleConfirm"
    @cancel="handleCancel"
    width="560px"
    placement="center"
  >
    <div class="email-notification-form">
      <!-- 知识库信息 -->
      <div class="form-item">
        <label class="form-label">{{ t('emailNotification.knowledgeBase') }}</label>
        <div class="kb-info">
          <t-tag theme="primary" variant="light">{{ kbName || '--' }}</t-tag>
        </div>
      </div>

      <!-- 收件人 -->
      <div class="form-item">
        <label class="form-label">
          {{ t('emailNotification.recipients') }}
          <span class="required">*</span>
        </label>
        <div class="recipient-input-row">
          <t-input
            v-model="recipientInput"
            :placeholder="t('emailNotification.recipientPlaceholder')"
            clearable
            @keydown.enter.prevent="addRecipient"
            @blur="addRecipient"
          />
          <t-button
            theme="default"
            variant="outline"
            size="medium"
            @click="addRecipient"
            :disabled="!recipientInput.trim()"
          >
            {{ t('emailNotification.addRecipient') }}
          </t-button>
        </div>
        <div class="recipient-tip">{{ t('emailNotification.recipientTip') }}</div>
        <div v-if="recipients.length" class="recipient-tags">
          <t-tag
            v-for="(email, idx) in recipients"
            :key="idx"
            theme="primary"
            variant="light-outline"
            closable
            size="medium"
            @close="removeRecipient(idx)"
          >
            {{ email }}
          </t-tag>
        </div>
      </div>

      <!-- 通知消息 -->
      <div class="form-item">
        <label class="form-label">
          {{ t('emailNotification.message') }}
          <span class="required">*</span>
        </label>
        <t-textarea
          v-model="message"
          :placeholder="t('emailNotification.messagePlaceholder')"
          :maxlength="500"
          :autosize="{ minRows: 2, maxRows: 4 }"
        />
      </div>

      <!-- 更新内容摘要（可选） -->
      <div class="form-item">
        <label class="form-label">
          {{ t('emailNotification.updateSummary') }}
          <span class="optional">({{ t('emailNotification.optional') }})</span>
        </label>
        <t-textarea
          v-model="updateSummary"
          :placeholder="t('emailNotification.updateSummaryPlaceholder')"
          :maxlength="1000"
          :autosize="{ minRows: 2, maxRows: 5 }"
        />
      </div>
    </div>
  </t-dialog>
</template>

<style scoped lang="less">
.email-notification-form {
  display: flex;
  flex-direction: column;
  gap: 18px;
  padding: 4px 0;

  .form-item {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .form-label {
    font-size: 13px;
    font-weight: 500;
    color: #333;

    .required {
      color: #e34d59;
      margin-left: 2px;
    }

    .optional {
      color: #999;
      font-weight: 400;
      font-size: 12px;
    }
  }

  .kb-info {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .recipient-input-row {
    display: flex;
    gap: 8px;
    align-items: center;

    .t-input {
      flex: 1;
    }
  }

  .recipient-tip {
    font-size: 12px;
    color: #999;
    line-height: 1.4;
  }

  .recipient-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    padding: 4px 0;
  }
}
</style>
