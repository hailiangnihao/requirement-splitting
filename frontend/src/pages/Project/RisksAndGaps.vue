<template>
  <div class="risks-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-main">
        <h2>风险与缺口分析</h2>
        <span class="sub-title">AI 根据当前项目物料与流转状态实时穿透计算</span>
      </div>
      <el-button type="primary" icon="Refresh" @click="recalculate" :loading="isCalculating">
        重新执行全量扫描
      </el-button>
    </div>

    <div v-loading="isCalculating" element-loading-text="AI 正在扫描全链路数据...">
      <!-- 整体健康度概览 -->
      <div class="health-overview">
        <div class="score-section">
          <el-progress type="dashboard" :percentage="score" :color="scoreColor" :width="100">
            <template #default="{ percentage }">
              <span class="percentage-value">{{ percentage }}</span>
              <span class="percentage-label">分</span>
            </template>
          </el-progress>
          <div class="health-desc">
            <h3>项目进度健康度：<span class="text-warning">{{ healthStatus }}</span></h3>
            <p>{{ summary }}</p>
          </div>
        </div>
      </div>

      <el-divider border-style="dashed" />

      <!-- 分区展示 -->
      <div class="risk-sections">
        <!-- 覆盖缺口区 -->
        <div class="section-column">
          <div class="section-title">
            <el-icon color="#E6A23C" :size="20"><Warning /></el-icon>
            <h3>覆盖缺口 (Coverage Gaps)</h3>
            <el-tag size="small" type="warning" round>{{ coverageGaps.length }}</el-tag>
          </div>
          
          <div class="card-list">
            <el-card v-for="gap in coverageGaps" :key="gap.id" class="risk-card gap-card" shadow="hover">
              <div class="card-header">
                <el-tag :type="getSeverityType(gap.severity)" size="small" effect="dark">{{ getSeverityText(gap.severity) }}缺口</el-tag>
                <span class="target-name">{{ gap.target }}</span>
              </div>
              <div class="card-desc">{{ gap.desc }}</div>
              
              <div class="ai-suggestion-box">
                <div class="ai-title"><el-icon><MagicStick /></el-icon> AI 处理建议</div>
                <div class="ai-content">{{ gap.aiSuggestion }}</div>
              </div>
              
              <div class="card-actions">
                <el-button type="primary" plain size="small" @click="handleAction(gap.actionText)">
                  {{ gap.actionText }}
                </el-button>
                <el-button link type="info" size="small">忽略</el-button>
              </div>
            </el-card>
          </div>
        </div>

        <!-- 阻塞与延期风险区 -->
        <div class="section-column">
          <div class="section-title">
            <el-icon color="#F56C6C" :size="20"><CircleCloseFilled /></el-icon>
            <h3>阻塞与延期风险 (Blocking Risks)</h3>
            <el-tag size="small" type="danger" round>{{ blockingRisks.length }}</el-tag>
          </div>
          
          <div class="card-list">
            <el-card v-for="risk in blockingRisks" :key="risk.id" class="risk-card block-card" shadow="hover">
              <div class="card-header">
                <el-tag :type="getSeverityType(risk.severity)" size="small" effect="dark">{{ getSeverityText(risk.severity) }}风险</el-tag>
                <span class="target-name">{{ risk.target }}</span>
              </div>
              <div class="card-desc">{{ risk.desc }}</div>
              
              <div class="ai-suggestion-box">
                <div class="ai-title"><el-icon><MagicStick /></el-icon> AI 协调建议</div>
                <div class="ai-content">{{ risk.aiSuggestion }}</div>
              </div>
              
              <div class="card-actions">
                <el-button type="danger" plain size="small" @click="handleAction(risk.actionText)">
                  {{ risk.actionText }}
                </el-button>
                <el-button link type="primary" size="small">查看详情</el-button>
              </div>
            </el-card>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import { ElMessage } from 'element-plus';
import { Warning, CircleCloseFilled, MagicStick, Refresh } from '@element-plus/icons-vue';
import { api } from '../../api/client';

const route = useRoute();
const isCalculating = ref(false);
const health = ref(null);

const scoreColor = [
  { color: '#f56c6c', percentage: 50 },
  { color: '#e6a23c', percentage: 80 },
  { color: '#67c23a', percentage: 100 },
];

const score = computed(() => health.value?.metrics?.base_score || 0);
const insight = computed(() => health.value?.insight || {});
const summary = computed(() => insight.value.executive_summary || '暂无健康度分析。');
const healthStatus = computed(() => score.value >= 80 ? '健康' : score.value >= 60 ? '关注' : '风险');
const coverageGaps = computed(() => {
  const metrics = health.value?.metrics;
  if (!metrics || metrics.untested_feature_count === 0) return [];
  return [{
    id: 'GAP-001',
    severity: 'high',
    target: '测试覆盖缺口',
    desc: `当前有 ${metrics.untested_feature_count} 个功能点没有测试用例。`,
    aiSuggestion: insight.value.action_items?.[0] || '建议补齐测试用例。',
    actionText: '查看测试用例'
  }];
});
const blockingRisks = computed(() => (insight.value.top_risks || []).map((risk, index) => ({
  id: `RSK-${index + 1}`,
  severity: index === 0 ? 'high' : 'medium',
  target: risk.title,
  desc: risk.description,
  aiSuggestion: insight.value.action_items?.[index] || '建议跟进处理。',
  actionText: '查看详情'
})));

const loadHealth = async () => {
  health.value = await api.getHealth(route.params.id);
};

onMounted(loadHealth);

// 辅助格式化函数
const getSeverityType = (s) => ({ 'critical': 'danger', 'high': 'warning', 'medium': 'info' }[s] || 'info');
const getSeverityText = (s) => ({ 'critical': '致命', 'high': '严重', 'medium': '一般' }[s] || s);

// 交互逻辑
const recalculate = () => {
  isCalculating.value = true;
  loadHealth().then(() => {
    isCalculating.value = false;
    ElMessage.success('全量扫描完成！数据已更新。');
  }).catch(error => {
    isCalculating.value = false;
    ElMessage.error(error.message || '健康度扫描失败');
  });
};

const handleAction = (actionName) => {
  ElMessage({
    message: `已触发操作：【${actionName}】，正在调用 AI 服务或发送通知...`,
    type: 'success',
    icon: MagicStick
  });
};
</script>

<style scoped>
.risks-page { background: #fff; padding: 24px; border-radius: 8px; box-shadow: 0 1px 4px rgba(0,0,0,0.05); min-height: calc(100vh - 110px); }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.header-main h2 { margin: 0 0 8px 0; font-size: 20px; color: #1f2f3d; }
.sub-title { font-size: 13px; color: #909399; }

.health-overview { display: flex; align-items: center; background: #f8f9fa; padding: 24px; border-radius: 8px; border: 1px solid #ebeef5; }
.score-section { display: flex; align-items: center; gap: 24px; }
.percentage-value { font-size: 28px; font-weight: bold; color: #303133; }
.percentage-label { font-size: 14px; color: #606266; }
.health-desc h3 { margin: 0 0 8px 0; font-size: 16px; color: #303133; }
.health-desc p { margin: 0; font-size: 14px; color: #606266; line-height: 1.6; }
.text-warning { color: #E6A23C; font-weight: bold; }

.risk-sections { display: grid; grid-template-columns: 1fr 1fr; gap: 32px; margin-top: 24px; }
.section-title { display: flex; align-items: center; gap: 8px; margin-bottom: 16px; padding-bottom: 12px; border-bottom: 2px solid #f0f2f5; }
.section-title h3 { margin: 0; font-size: 16px; color: #303133; }

.card-list { display: flex; flex-direction: column; gap: 16px; }
.risk-card { border-radius: 6px; border: none; background: #fff; box-shadow: 0 2px 8px rgba(0,0,0,0.04); transition: transform 0.2s; }
.risk-card:hover { transform: translateY(-2px); box-shadow: 0 4px 12px rgba(0,0,0,0.08); }

/* 不同风险类型的侧边标识线 */
.gap-card { border-left: 4px solid #E6A23C; }
.block-card { border-left: 4px solid #F56C6C; }

.risk-card :deep(.el-card__body) { padding: 16px; }
.card-header { display: flex; align-items: center; gap: 12px; margin-bottom: 12px; }
.target-name { font-weight: bold; color: #303133; font-size: 15px; }
.card-desc { font-size: 13px; color: #606266; line-height: 1.6; margin-bottom: 16px; }

.ai-suggestion-box { background: #f9f0ff; border-radius: 4px; padding: 12px; margin-bottom: 16px; border: 1px dashed #d3adf7; }
.ai-title { font-size: 13px; font-weight: bold; color: #722ed1; display: flex; align-items: center; gap: 6px; margin-bottom: 6px; }
.ai-content { font-size: 13px; color: #531dab; line-height: 1.5; }

.card-actions { display: flex; justify-content: flex-end; gap: 12px; }
</style>
