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
          <el-progress type="dashboard" :percentage="72" :color="scoreColor" :width="100">
            <template #default="{ percentage }">
              <span class="percentage-value">{{ percentage }}</span>
              <span class="percentage-label">分</span>
            </template>
          </el-progress>
          <div class="health-desc">
            <h3>项目进度健康度：<span class="text-warning">关注</span></h3>
            <p>经 AI 扫描，当前项目存在 <strong>2</strong> 个严重覆盖缺口，<strong>2</strong> 个阻塞性风险。建议优先处理验收阻塞与核心用例缺失问题。</p>
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
import { ref } from 'vue';
import { ElMessage } from 'element-plus';
import { Warning, CircleCloseFilled, MagicStick, Refresh } from '@element-plus/icons-vue';

const isCalculating = ref(false);

const scoreColor = [
  { color: '#f56c6c', percentage: 50 },
  { color: '#e6a23c', percentage: 80 },
  { color: '#67c23a', percentage: 100 },
];

// 模拟：覆盖缺口数据 (如测试覆盖、验收覆盖)
const coverageGaps = ref([
  { 
    id: 'GAP-001', severity: 'high', target: '功能点：飞书扫码登录', 
    desc: '该核心功能点目前没有任何测试用例关联，且关联的 2 个开发任务均已处于“开发中”，存在极大的漏测风险。', 
    aiSuggestion: '建议立即由 AI 根据需求草稿生成补充测试用例，并提交给测试人员复核。', 
    actionText: 'AI 一键生成用例' 
  },
  { 
    id: 'GAP-002', severity: 'medium', target: '里程碑：V1.0 基础功能上线', 
    desc: '该里程碑下存在 3 个前端展示相关的任务尚未关联任何明确的“验收检查项”。', 
    aiSuggestion: '建议让 AI 读取相关功能点并自动提取业务验收标准。', 
    actionText: 'AI 补充验收项' 
  }
]);

// 模拟：阻塞与延期风险数据 (如缺陷阻塞、进度超期)
const blockingRisks = ref([
  { 
    id: 'RSK-001', severity: 'critical', target: '缺陷：BUG-101 密码错误无弹窗', 
    desc: '该缺陷严重级别为“高”，且标记为【阻塞验收】，目前一直处于“待修复”状态，直接影响后续的 V1.0 回归测试。', 
    aiSuggestion: '建议提升该缺陷的优先级，并自动向当前处理人【张三】发送飞书/邮件催办提醒。', 
    actionText: '一键自动催办' 
  },
  { 
    id: 'RSK-002', severity: 'high', target: '任务：对接飞书 OAuth2.0', 
    desc: '该任务已超过计划完成时间 2 天，且作为“前端集成组件”任务的前置依赖，将导致相关链路整体延期。', 
    aiSuggestion: 'AI 分析该任务涉及外部依赖，建议与研发负责人协调，或重新评估关联里程碑的时间。', 
    actionText: '调整关联排期' 
  }
]);

// 辅助格式化函数
const getSeverityType = (s) => ({ 'critical': 'danger', 'high': 'warning', 'medium': 'info' }[s] || 'info');
const getSeverityText = (s) => ({ 'critical': '致命', 'high': '严重', 'medium': '一般' }[s] || s);

// 交互逻辑
const recalculate = () => {
  isCalculating.value = true;
  setTimeout(() => {
    isCalculating.value = false;
    ElMessage.success('全量扫描完成！数据已更新。');
  }, 2500);
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