<template>
  <div class="knowledge-push-container">
    <!-- 头部：仅标题与副标题 -->
    <div class="header">
      <div class="header-title">
        <h2>{{ $t('knowledgePush.title') }}</h2>
        <p class="header-subtitle">{{ $t('knowledgePush.subtitle') }}</p>
      </div>
    </div>

    <!-- 左侧菜单 + 主内容 -->
    <div class="knowledge-push-body">
      <ListSpaceSidebar
        v-model="spaceSelection"
        mode="resource"
        :count-all="allKBsCount"
        :count-mine="myKBsCount"
        :count-by-org="kbsByOrg"
      />
      <div class="knowledge-push-main">
        <t-steps :current="step" class="kp-steps">
          <t-step :title="$t('knowledgePush.step1')" />
          <t-step :title="$t('knowledgePush.step2')" />
          <t-step :title="$t('knowledgePush.step3')" />
        </t-steps>

        <div v-if="step === 0" class="kp-step-content">
          <!-- Step 1: 选择空间和知识库 -->
          <div class="kp-section">
            <label class="kp-label">{{ $t('knowledgePush.selectSpace') }}</label>
            <t-select v-model="selectedSpaceId" :loading="loadingSpaces" :placeholder="$t('knowledgePush.selectSpace')">
              <t-option v-for="space in spaces" :key="space.id" :value="space.id" :label="space.name" />
            </t-select>
          </div>
          <div class="kp-section">
            <label class="kp-label">{{ $t('knowledgePush.selectKB') }}</label>
            <t-select v-model="selectedKbId" :loading="loadingKbs" :placeholder="$t('knowledgePush.selectKB')" :disabled="!selectedSpaceId">
              <t-option v-for="kb in kbs" :key="kb.knowledge_base.id" :value="kb.knowledge_base.id" :label="kb.knowledge_base.name" />
            </t-select>
          </div>
          <div class="kp-actions">
            <t-button theme="primary" :disabled="!selectedKbId" @click="nextStep">{{ $t('knowledgePush.next') }}</t-button>
          </div>
        </div>
        <div v-else-if="step === 1" class="kp-step-content">
          <!-- Step 2: 选择收件人 -->
          <div class="kp-section">
            <label class="kp-label">{{ $t('knowledgePush.recipients') }}</label>
            <div class="kp-recipient-row">
              <t-input
                v-model="recipientInput"
                :placeholder="$t('knowledgePush.recipientPlaceholder')"
                clearable
                @keydown.enter.prevent="addRecipient"
                @blur="addRecipient"
                style="max-width: 320px;"
              />
              <t-button
                theme="default"
                variant="outline"
                size="medium"
                @click="addRecipient"
                :disabled="!recipientInput.trim()"
              >
                {{ $t('knowledgePush.addRecipient') }}
              </t-button>
              <t-button
                theme="default"
                size="medium"
                @click="clearAllRecipients"
                v-if="recipients.length > 0"
                style="margin-left: 8px;"
              >
                {{ $t('knowledgePush.clearAll') }}
              </t-button>
            </div>
            <div class="kp-recipient-tip">{{ $t('knowledgePush.recipientTip') }}</div>
            <div v-if="recipients.length" class="kp-recipient-tags">
              <t-tag
                v-for="(email, idx) in recipients"
                :key="email + idx"
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
          <div class="kp-section">
            <div class="kp-quick-select">
              <span class="kp-label">{{ $t('knowledgePush.quickSelect') }}</span>
              <t-button size="small" @click="openPlatformUsersDialog">{{ $t('knowledgePush.platformUsers') }}</t-button>
              <t-button size="small" @click="showSpaceMembers = true" :disabled="!selectedSpaceId">{{ $t('knowledgePush.spaceMembers') }}</t-button>
              <t-button size="small" @click="addAllMembers" :disabled="!selectedSpaceId || !spaceMembers.length">{{ $t('knowledgePush.sendToAllMembers') }}</t-button>
            </div>
          </div>
          <!-- 平台用户弹窗 -->
          <t-dialog v-model:visible="showPlatformUsers" :header="$t('knowledgePush.platformUsers')" width="500px" placement="center">
            <div>
              <t-input v-model="userSearch" :placeholder="$t('knowledgePush.searchUsersPlaceholder')" @input="searchUsers" clearable style="margin-bottom: 12px;" />
              <div v-if="loadingUsers">{{ $t('knowledgePush.loadingUsers') }}</div>
              <div v-else-if="platformUsers.length === 0">{{ $t('knowledgePush.noUsersFound') }}</div>
              <div v-else class="kp-user-list">
                <div v-for="user in platformUsers" :key="user.email" class="kp-user-item">
                  <span>{{ user.username }} ({{ user.email }})</span>
                  <t-button size="small" @click="addRecipientEmail(user.email)">{{ $t('knowledgePush.addRecipient') }}</t-button>
                </div>
              </div>
            </div>
          </t-dialog>
          <!-- 空间成员弹窗 -->
          <t-dialog v-model:visible="showSpaceMembers" :header="$t('knowledgePush.spaceMembers')" width="500px" placement="center">
            <div>
              <div v-if="loadingMembers">{{ $t('knowledgePush.loadingMembers') }}</div>
              <div v-else-if="spaceMembers.length === 0">{{ $t('knowledgePush.noUsersFound') }}</div>
              <div v-else class="kp-user-list">
                <div v-for="member in spaceMembers" :key="member.email" class="kp-user-item">
                  <span>{{ member.username }} ({{ member.email }})</span>
                  <t-button size="small" @click="addRecipientEmail(member.email)">{{ $t('knowledgePush.addRecipient') }}</t-button>
                </div>
              </div>
            </div>
          </t-dialog>
          <div class="kp-actions">
            <t-button variant="outline" @click="step = 0">{{ $t('knowledgePush.prev') }}</t-button>
            <t-button theme="primary" :disabled="recipients.length === 0" @click="nextStep">{{ $t('knowledgePush.next') }}</t-button>
          </div>
        </div>
        <div v-else-if="step === 2" class="kp-step-content">
          <!-- Step 3: 编写消息并发送 -->
          <div class="kp-section">
            <label class="kp-label">{{ $t('knowledgePush.message') }}<span class="required">*</span></label>
            <t-textarea
              v-model="message"
              :placeholder="$t('knowledgePush.messagePlaceholder')"
              :maxlength="500"
              :autosize="{ minRows: 2, maxRows: 4 }"
            />
          </div>
          <div class="kp-section">
            <label class="kp-label">{{ $t('knowledgePush.updateSummary') }}<span class="optional">({{ $t('knowledgePush.optional') }})</span></label>
            <t-textarea
              v-model="updateSummary"
              :placeholder="$t('knowledgePush.updateSummaryPlaceholder')"
              :maxlength="1000"
              :autosize="{ minRows: 2, maxRows: 5 }"
            />
          </div>
          <div class="kp-actions">
            <t-button variant="outline" @click="step = 1">{{ $t('knowledgePush.prev') }}</t-button>
            <t-button theme="primary" :loading="sending" :disabled="!canSubmit" @click="handleSend">{{ $t('knowledgePush.send') }}</t-button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { MessagePlugin } from 'tdesign-vue-next';
import ListSpaceSidebar from '@/components/ListSpaceSidebar.vue';
import { listMyOrganizations, listOrganizationSharedKnowledgeBases, listMembers } from '@/api/organization/index';
import { searchPlatformUsers, sendKBUpdateNotification, listKnowledgeBases } from '@/api/knowledge-base/index';

const { t } = useI18n();

// 左侧边栏相关
const spaceSelection = ref('all');
const allMyOrganizations = ref<any[]>([]);
const allMyKnowledgeBases = ref<any[]>([]);
const myKBsCount = computed(() => allMyKnowledgeBases.value.length);
const allKBsCount = computed(() => {
  // 计算所有可访问的知识库总数
  let total = allMyKnowledgeBases.value.length;
  Object.values(kbsByOrg.value).forEach((count: any) => {
    total += count;
  });
  return total;
});
const kbsByOrg = computed(() => {
  // 统计各组织下的知识库数量
  const counts: Record<string, number> = {};
  allMyOrganizations.value.forEach((org: any) => {
    counts[org.id] = org._kbCount || 0;
  });
  return counts;
});

// 根据左侧菜单选择获取对应的知识库
const availableKBs = computed(() => {
  if (spaceSelection.value === 'all') {
    return allMyKnowledgeBases.value;
  }
  if (spaceSelection.value === 'mine') {
    return allMyKnowledgeBases.value;
  }
  // 共享空间
  return allMyKnowledgeBases.value.filter(
    (kb: any) => kb.organization_id === spaceSelection.value || kb.share_organization_id === spaceSelection.value
  );
});

const step = ref(0);
const spaces = ref<any[]>([]);
const loadingSpaces = ref(false);
const selectedSpaceId = ref('');
const kbs = ref<any[]>([]);
const loadingKbs = ref(false);
const selectedKbId = ref('');

// Step 2: 收件人
const recipientInput = ref('');
const recipients = ref<string[]>([]);
const showPlatformUsers = ref(false);
const showSpaceMembers = ref(false);
const userSearch = ref('');
const platformUsers = ref<any[]>([]);
const loadingUsers = ref(false);
const spaceMembers = ref<any[]>([]);
const loadingMembers = ref(false);

// Step 3: 消息与发送
const message = ref('');
const updateSummary = ref('');
const sending = ref(false);

const canSubmit = computed(() => {
  return recipients.value.length > 0 && message.value.trim().length > 0;
});

// 加载全部空间和知识库
const loadAllData = async () => {
  loadingSpaces.value = true;
  try {
    const res = await listMyOrganizations();
    allMyOrganizations.value = res.data?.organizations || [];
    spaces.value = allMyOrganizations.value; // 同时更新 spaces 供模板使用
    
    // 为每个组织加载知识库
    for (const org of allMyOrganizations.value) {
      const kbRes = await listOrganizationSharedKnowledgeBases(org.id);
      org._kbCount = (kbRes.data || []).length;
    }
  } catch (e) {
    allMyOrganizations.value = [];
    spaces.value = [];
  } finally {
    loadingSpaces.value = false;
  }

  // 加载我的知识库
  try {
    const kbRes = await listKnowledgeBases();
    allMyKnowledgeBases.value = kbRes.data || [];
  } catch (e) {
    allMyKnowledgeBases.value = [];
  }
};

// 加载空间列表
const loadSpaces = async () => {
  loadingSpaces.value = true;
  try {
    const res = await listMyOrganizations();
    spaces.value = res.data?.organizations || [];
  } catch (e) {
    spaces.value = [];
  } finally {
    loadingSpaces.value = false;
  }
};

// 加载选中空间下的知识库
const loadKbs = async (orgId: string) => {
  loadingKbs.value = true;
  try {
    const res = await listOrganizationSharedKnowledgeBases(orgId);
    kbs.value = res.data || [];
  } catch (e) {
    kbs.value = [];
  } finally {
    loadingKbs.value = false;
  }
};

// 加载空间成员
const loadSpaceMembers = async (orgId: string) => {
  loadingMembers.value = true;
  try {
    const res = await listMembers(orgId);
    spaceMembers.value = res.data?.members || [];
  } catch (e) {
    spaceMembers.value = [];
  } finally {
    loadingMembers.value = false;
  }
};

// 搜索平台用户（空字符串时返回全部用户）
const searchUsers = async () => {
  loadingUsers.value = true;
  try {
    const res = await searchPlatformUsers(userSearch.value.trim(), 100);
    platformUsers.value = res.data || [];
  } catch (e) {
    platformUsers.value = [];
  } finally {
    loadingUsers.value = false;
  }
};

// 打开平台用户弹窗并加载全量用户
const openPlatformUsersDialog = () => {
  showPlatformUsers.value = true;
  userSearch.value = '';
  searchUsers();
};

// watch 选中空间，自动加载成员
watch(selectedSpaceId, (val) => {
  selectedKbId.value = '';
  if (val) {
    loadKbs(val);
    loadSpaceMembers(val);
  } else {
    kbs.value = [];
    spaceMembers.value = [];
  }
});

const nextStep = () => {
  if (step.value === 0 && selectedKbId.value) step.value = 1;
  else if (step.value === 1 && recipients.value.length > 0) step.value = 2;
};

// 收件人相关
const isValidEmail = (email: string): boolean => {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim());
};
const addRecipient = () => {
  const raw = recipientInput.value.trim();
  if (!raw) return;
  const emails = raw.split(/[,;，；\s]+/).filter(Boolean);
  for (const email of emails) {
    const trimmed = email.trim();
    if (!isValidEmail(trimmed)) continue;
    if (!recipients.value.includes(trimmed)) recipients.value.push(trimmed);
  }
  recipientInput.value = '';
};
const addRecipientEmail = (email: string) => {
  if (isValidEmail(email) && !recipients.value.includes(email)) {
    recipients.value.push(email);
  }
};
const removeRecipient = (idx: number) => {
  recipients.value.splice(idx, 1);
};
const clearAllRecipients = () => {
  recipients.value = [];
};
// 一键添加所有成员邮箱
const addAllMembers = () => {
  if (!spaceMembers.value.length) return;
  let added = 0;
  for (const member of spaceMembers.value) {
    if (member.email && isValidEmail(member.email) && !recipients.value.includes(member.email)) {
      recipients.value.push(member.email);
      added++;
    }
  }
};

const handleSend = async () => {
  if (!canSubmit.value) return;
  sending.value = true;
  try {
    const res: any = await sendKBUpdateNotification({
      knowledge_base_id: selectedKbId.value,
      recipients: recipients.value,
      message: message.value.trim(),
      update_summary: updateSummary.value.trim() || undefined,
    });
    const data = res?.data;
    if (data) {
      if (data.fail_count === 0) {
        MessagePlugin.success(t('knowledgePush.sendSuccess', { count: data.success_count }));
      } else if (data.success_count > 0) {
        MessagePlugin.warning(t('knowledgePush.sendPartial', { success: data.success_count, fail: data.fail_count }));
      } else {
        MessagePlugin.error(t('knowledgePush.sendAllFailed'));
      }
    } else {
      MessagePlugin.success(t('knowledgePush.sendSuccess', { count: recipients.value.length }));
    }
    // 重置表单，回到第一步
    step.value = 0;
    recipients.value = [];
    message.value = '';
    updateSummary.value = '';
  } catch (error: any) {
    const errorMsg = error?.message || error?.error?.message || t('knowledgePush.sendFailed');
    MessagePlugin.error(errorMsg);
  } finally {
    sending.value = false;
  }
};

onMounted(() => {
  loadAllData();
});
</script>

<style scoped lang="less">
.knowledge-push-container {
  padding: 24px 32px;
  margin: 0 16px 0 4px;
  height: calc(100vh);
  box-sizing: border-box;
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

// 头部：标题区
.header {
  display: flex;
  align-items: center;
  margin-bottom: 20px;

  .header-title {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  h2 {
    margin: 0;
    color: #000000e6;
    font-family: "PingFang SC";
    font-size: 24px;
    font-weight: 600;
    line-height: 32px;
  }
}

.header-subtitle {
  margin: 0;
  color: #00000099;
  font-family: "PingFang SC";
  font-size: 14px;
  font-weight: 400;
  line-height: 20px;
}

// 主体：左侧菜单 + 右侧内容
.knowledge-push-body {
  display: flex;
  flex: 1;
  min-height: 0;
  background: #fafbfc;
  border: 1px solid #e7ebf0;
  border-radius: 10px;
  overflow: hidden;
}

.knowledge-push-main {
  flex: 1;
  min-width: 0;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 12px;
  background: #fafbfc;
}

// 步骤条
.kp-steps {
  margin-bottom: 24px;
}

// 内容区
.kp-step-content {
  background-color: #fff;
  border-radius: 4px;
  padding: 24px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
}

.kp-section {
  margin-bottom: 24px;

  &:last-child {
    margin-bottom: 0;
  }
}

.kp-label {
  font-size: 14px;
  font-weight: 500;
  margin-bottom: 8px;
  display: block;
  color: #0a0e27;
}

.kp-actions {
  margin-top: 24px;
  text-align: right;
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

.required {
  color: #e34d59;
  margin-left: 4px;
}

.optional {
  color: #999;
  font-weight: 400;
  font-size: 12px;
  margin-left: 4px;
}

/* 收件人选择相关样式 */
.kp-recipient-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.kp-recipient-tip {
  font-size: 12px;
  color: #999;
  line-height: 1.4;
  margin-top: 6px;
}

.kp-recipient-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  padding: 8px 0;
}

.kp-quick-select {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 12px;
  padding: 12px;
  background: #f5f7fa;
  border-radius: 4px;

  span.kp-label {
    margin-bottom: 0;
  }
}

.kp-user-list {
  max-height: 260px;
  overflow-y: auto;
}

.kp-user-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 0;
  border-bottom: 1px solid #f0f0f0;

  &:last-child {
    border-bottom: none;
  }

  span {
    flex: 1;
    font-size: 14px;
  }
}
</style>
