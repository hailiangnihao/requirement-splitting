<template>
  <div class="split-container">
    <!-- 左侧：原始需求区 -->
    <div class="left-panel">
      <div class="panel-header">
        <h3>原始需求</h3>
        <el-button type="primary" plain size="small">上传文档 (Word/PDF)</el-button>
      </div>
      <div class="input-wrapper">
        <el-input
          v-model="rawRequirement"
          type="textarea"
          placeholder="请输入您的原始需求描述，例如：&#10;我们需要一个用户登录功能，支持账号密码登录，登录后跳转到项目列表。如果没有账号可以进行注册..."
          class="req-input"
        />
      </div>
      <div class="panel-footer">
        <div class="ai-config">
          <span class="label">拆分策略：</span>
          <el-radio-group v-model="splitStrategy" size="small">
            <el-radio-button label="stage1">一阶段 (模块/里程碑)</el-radio-button>
            <el-radio-button label="stage2">全量深度拆分</el-radio-button>
          </el-radio-group>
        </div>
        <el-button type="primary" size="large" @click="handleAiSplit" :loading="isSplitting">
          <el-icon class="el-icon--left"><MagicStick /></el-icon>
          AI 智能拆分
        </el-button>
      </div>
    </div>

    <!-- 右侧：AI 拆分草稿 (结构化树) -->
    <div class="right-panel">
      <div class="panel-header">
        <div class="title-with-badge">
          <h3>AI 拆分草稿</h3>
          <el-tag v-if="hasResult" type="warning" size="small" effect="light">草稿未发布</el-tag>
        </div>
        <div class="actions">
          <el-button size="small" @click="resetTree" :disabled="isSplitting">重置</el-button>
          <el-button type="success" size="small" :disabled="!hasResult" @click="publishPlan">
            确认并发布正式计划
          </el-button>
        </div>
      </div>

      <div class="tree-container" v-loading="isSplitting" element-loading-text="AI 正在深度思考和拆分中...">
        <el-empty v-if="!hasResult && !isSplitting" description="暂无拆分结果，请在左侧输入需求并点击拆分" />
        
        <el-tree
          v-if="hasResult"
          :data="treeData"
          :props="defaultProps"
          node-key="id"
          default-expand-all
          :expand-on-click-node="false"
        >
          <template #default="{ node, data }">
            <div class="custom-tree-node">
              <div class="node-main">
                <el-tag :type="getTagType(data.type)" size="small" class="node-tag">
                  {{ data.type }}
                </el-tag>
                <span class="node-label">{{ node.label }}</span>
              </div>
              <div class="node-extra">
                <span v-if="data.desc" class="node-desc" :title="data.desc">{{ data.desc }}</span>
                <!-- 草稿状态下的人工微调操作 -->
                <div class="node-actions">
                  <el-button link type="primary" size="small">编辑</el-button>
                  <el-button link type="danger" size="small">删除</el-button>
                </div>
              </div>
            </div>
          </template>
        </el-tree>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { ElMessage } from 'element-plus';
import { MagicStick } from '@element-plus/icons-vue';

const rawRequirement = ref('');
const splitStrategy = ref('stage2');
const isSplitting = ref(false);
const hasResult = ref(false);
const treeData = ref([]);

const defaultProps = { children: 'children', label: 'label' };

// 模拟 AI 拆分过程
const handleAiSplit = () => {
  if (!rawRequirement.value.trim()) {
    ElMessage.warning('请先输入原始需求');
    return;
  }
  
  isSplitting.value = true;
  hasResult.value = false;

  // 模拟请求延迟，展示 AI 思考态
  setTimeout(() => {
    treeData.value = [
      {
        id: 1, label: '用户认证模块', type: '模块',
        children: [
          {
            id: 11, label: '账号密码登录', type: '功能点', desc: '支持用户名/手机号+密码登录，包含错误重试限制',
            children: [
              { id: 111, label: '开发：登录接口开发及鉴权', type: '开发任务' },
              { id: 112, label: '开发：前端登录页面及表单校验', type: '开发任务' },
              { id: 113, label: '测试：登录异常场景(错密/空值)测试', type: '测试用例' },
              { id: 114, label: '验收：有效用户能够正常登录系统并保持会话', type: '验收项' },
            ]
          }
        ]
      },
      {
        id: 2, label: '项目管理模块', type: '模块',
        children: [
          {
            id: 21, label: '项目列表与搜索', type: '功能点', desc: '支持按状态、负责人筛选项目',
            children: [
              { id: 211, label: '开发：列表查询接口集成', type: '开发任务' }
            ]
          }
        ]
      }
    ];
    isSplitting.value = false;
    hasResult.value = true;
    ElMessage.success('🎉 AI 拆分完成，已生成拆分草稿，请人工确认！');
  }, 2500);
};

const resetTree = () => {
  treeData.value = [];
  hasResult.value = false;
};

const publishPlan = () => {
  // 实际项目中这里会将草稿数据提交给后端生成正式版本
  ElMessage.success('✅ 正式计划发布成功！可在任务看板和测试页查看');
};

// 根据节点类型返回不同的标签颜色
const getTagType = (type) => {
  const map = {
    '模块': 'primary',
    '功能点': 'success',
    '开发任务': 'warning',
    '测试用例': 'info',
    '验收项': 'danger'
  };
  return map[type] || 'info';
};
</script>

<style scoped>
.split-container {
  display: flex;
  gap: 24px;
  /* 撑满外层 page-container 的高度 */
  height: calc(100vh - 110px); 
}
.left-panel, .right-panel {
  flex: 1;
  background: #fff;
  border-radius: 8px;
  padding: 20px;
  display: flex;
  flex-direction: column;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);
  min-width: 0; /* 防止 flex 子项溢出 */
}
.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  border-bottom: 1px solid #f0f2f5;
  padding-bottom: 12px;
}
.panel-header h3 { margin: 0; font-size: 16px; color: #1f2f3d; }
.title-with-badge { display: flex; align-items: center; gap: 8px; }
.input-wrapper {
  flex: 1;
  display: flex;
  margin-bottom: 16px;
}
/* 深度选择器让输入框高度撑满 */
.input-wrapper :deep(.el-textarea) { flex: 1; display: flex; }
.input-wrapper :deep(.el-textarea__inner) { flex: 1; resize: none; font-size: 14px; }
.panel-footer { display: flex; justify-content: space-between; align-items: center; }
.ai-config .label { font-size: 13px; color: #606266; margin-right: 8px; }

.tree-container {
  flex: 1;
  overflow-y: auto;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  padding: 16px;
}
.custom-tree-node {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 14px;
  padding-right: 8px;
  width: 100%;
}
.node-main { display: flex; align-items: center; }
.node-tag { margin-right: 8px; width: 66px; text-align: center; }
.node-label { font-weight: 500; color: #303133; }
.node-extra { display: flex; align-items: center; }
.node-desc {
  margin-right: 16px;
  font-size: 12px;
  color: #909399;
  max-width: 200px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.node-actions { display: none; }
.custom-tree-node:hover .node-actions { display: block; }
</style>