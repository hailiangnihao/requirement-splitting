<template>
  <div class="dashboard-page">
    <!-- 顶部：项目标题与 AI 健康度状态 -->
    <div class="page-header">
      <div class="header-left">
        <h2>{{ project?.name || '项目总览' }}</h2>
        <el-tag type="warning" effect="dark" size="large" class="health-tag">
          <el-icon><Warning /></el-icon> 进度健康度：{{ healthLabel }}
        </el-tag>
      </div>
      <div class="header-right">
        <span class="update-time">最后更新：10分钟前</span>
        <el-button type="primary">发布周报</el-button>
      </div>
    </div>

    <!-- 核心指标卡片 -->
    <div class="metrics-row">
      <el-card shadow="hover" class="metric-card">
        <div class="metric-title">任务完成率</div>
        <div class="metric-value">{{ taskRate }}<span class="unit">%</span></div>
        <div class="metric-footer">
          总任务: {{ health?.metrics?.dev_task_total || 0 }} <el-divider direction="vertical" /> 已完成: {{ health?.metrics?.dev_task_done || 0 }}
        </div>
      </el-card>
      <el-card shadow="hover" class="metric-card">
        <div class="metric-title">测试覆盖率</div>
        <div class="metric-value">{{ coverageRate }}<span class="unit">%</span></div>
        <div class="metric-footer">
          未覆盖功能点: {{ health?.metrics?.untested_feature_count || 0 }}
        </div>
      </el-card>
      <el-card shadow="hover" class="metric-card">
        <div class="metric-title">活跃缺陷</div>
        <div class="metric-value text-danger">{{ health?.metrics?.active_defects || 0 }}<span class="unit">个</span></div>
        <div class="metric-footer">
          当前未关闭缺陷数量
        </div>
      </el-card>
      <el-card shadow="hover" class="metric-card">
        <div class="metric-title">变更冲击</div>
        <div class="metric-value text-warning">{{ health?.metrics?.recent_changes || 0 }}<span class="unit">项</span></div>
        <div class="metric-footer">
          待处理的需求变更影响范围
        </div>
      </el-card>
    </div>

    <!-- 图表区 -->
    <div class="charts-row">
      <el-card shadow="hover" class="chart-card">
        <template #header>
          <div class="card-header">
            <span>任务进度分布</span>
          </div>
        </template>
        <div ref="taskChartRef" class="chart-container"></div>
      </el-card>

      <el-card shadow="hover" class="chart-card">
        <template #header>
          <div class="card-header">
            <span>缺陷趋势与状态</span>
          </div>
        </template>
        <div ref="bugChartRef" class="chart-container"></div>
      </el-card>
    </div>

    <!-- 底部：AI 洞察与里程碑 -->
    <div class="bottom-row">
      <!-- AI 风险评估卡片 -->
      <el-card shadow="hover" class="ai-insight-card">
        <template #header>
          <div class="card-header ai-header">
            <el-icon><MagicStick /></el-icon>
            <span>AI 风险与进度洞察</span>
          </div>
        </template>
        <div class="insight-content">
          <el-alert
            title="AI 洞察"
            type="error"
            :description="healthInsight.executive_summary || '暂无风险洞察。'"
            show-icon
            :closable="false"
            class="insight-alert"
          />
          <el-alert
            title="重点风险"
            type="warning"
            :description="firstRisk"
            show-icon
            :closable="false"
            class="insight-alert"
          />
          <el-alert
            title="下一步建议"
            type="success"
            :description="firstAction"
            show-icon
            :closable="false"
            class="insight-alert"
          />
        </div>
      </el-card>

      <!-- 里程碑时间轴 -->
      <el-card shadow="hover" class="timeline-card">
        <template #header>
          <div class="card-header">
            <span>项目里程碑</span>
          </div>
        </template>
        <el-timeline>
          <el-timeline-item timestamp="2023-10-01" placement="top" type="success">
            <el-card shadow="never" class="timeline-inner-card">
              <h4>立项与需求确认</h4>
              <p>完成需求文档拆分并发布 V1.0 正式计划</p>
            </el-card>
          </el-timeline-item>
          <el-timeline-item timestamp="2023-10-20" placement="top" type="primary">
            <el-card shadow="never" class="timeline-inner-card">
              <h4>V1.0 基础功能提测</h4>
              <p>研发进度 100%，目前测试进度 60%</p>
            </el-card>
          </el-timeline-item>
          <el-timeline-item timestamp="2023-11-05" placement="top" color="#e4e7ed">
            <el-card shadow="never" class="timeline-inner-card">
              <h4>V1.0 灰度发布与验收</h4>
              <p>待验收功能点：24 个</p>
            </el-card>
          </el-timeline-item>
        </el-timeline>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, markRaw, nextTick } from 'vue';
import { useRoute } from 'vue-router';
import * as echarts from 'echarts';
import { Warning, MagicStick } from '@element-plus/icons-vue';
import { api } from '../../api/client';

const route = useRoute();
const taskChartRef = ref(null);
const bugChartRef = ref(null);
const project = ref(null);
const health = ref(null);

let taskChart = null;
let bugChart = null;

onMounted(() => {
  loadOverview();
  window.addEventListener('resize', handleResize);
});

onUnmounted(() => {
  if (taskChart) taskChart.dispose();
  if (bugChart) bugChart.dispose();
  window.removeEventListener('resize', handleResize);
});

const handleResize = () => {
  if (taskChart) taskChart.resize();
  if (bugChart) bugChart.resize();
};

const healthInsight = computed(() => health.value?.insight || {});
const healthLabel = computed(() => {
  const score = health.value?.metrics?.base_score ?? 0;
  if (score >= 80) return '健康';
  if (score >= 60) return '关注';
  return '风险';
});
const taskRate = computed(() => {
  const total = health.value?.metrics?.dev_task_total || 0;
  const done = health.value?.metrics?.dev_task_done || 0;
  return total ? Math.round((done / total) * 100) : 0;
});
const coverageRate = computed(() => {
  const total = health.value?.metrics?.feature_point_count || 0;
  const untested = health.value?.metrics?.untested_feature_count || 0;
  return total ? Math.round(((total - untested) / total) * 100) : 0;
});
const firstRisk = computed(() => healthInsight.value.top_risks?.[0]?.description || '暂无重点风险。');
const firstAction = computed(() => healthInsight.value.action_items?.[0] || '暂无下一步建议。');

const loadOverview = async () => {
  project.value = await api.getProject(route.params.id);
  health.value = await api.getHealth(route.params.id);
  await nextTick();
  initTaskChart();
  initBugChart();
};

// 初始化任务环形图
const initTaskChart = () => {
  taskChart = markRaw(echarts.init(taskChartRef.value));
  const option = {
    tooltip: { trigger: 'item' },
    legend: { top: '5%', left: 'center' },
    color: ['#909399', '#409EFF', '#E6A23C', '#F56C6C', '#67C23A'],
    series: [
      {
        name: '任务分布',
        type: 'pie',
        radius: ['40%', '70%'],
        avoidLabelOverlap: false,
        itemStyle: { borderRadius: 10, borderColor: '#fff', borderWidth: 2 },
        label: { show: false, position: 'center' },
        emphasis: {
          label: { show: true, fontSize: 18, fontWeight: 'bold' }
        },
        labelLine: { show: false },
        data: [
          { value: health.value?.metrics?.dev_task_total || 0, name: '总任务' },
          { value: health.value?.metrics?.dev_task_done || 0, name: '已完成' }
        ]
      }
    ]
  };
  taskChart.setOption(option);
};

// 初始化缺陷柱状图
const initBugChart = () => {
  bugChart = markRaw(echarts.init(bugChartRef.value));
  const option = {
    tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { type: 'category', data: ['致命', '严重', '一般', '轻微'] },
    yAxis: { type: 'value' },
    series: [
      {
        name: '缺陷数量',
        type: 'bar',
        barWidth: '40%',
        itemStyle: {
          color: function(params) {
            const colorList = ['#722ed1', '#F56C6C', '#E6A23C', '#909399'];
            return colorList[params.dataIndex];
          },
          borderRadius: [4, 4, 0, 0]
        },
        data: [0, health.value?.metrics?.active_defects || 0, 0, 0]
      }
    ]
  };
  bugChart.setOption(option);
};
</script>

<style scoped>
.dashboard-page { padding: 20px; min-height: calc(100vh - 60px); }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.header-left { display: flex; align-items: center; gap: 16px; }
.header-left h2 { margin: 0; font-size: 22px; color: #1f2f3d; }
.health-tag { font-size: 14px; font-weight: 500; }
.header-right { display: flex; align-items: center; gap: 16px; }
.update-time { font-size: 13px; color: #909399; }

.metrics-row { display: grid; grid-template-columns: repeat(4, 1fr); gap: 20px; margin-bottom: 24px; }
.metric-card { border-radius: 8px; border: none; }
.metric-card :deep(.el-card__body) { padding: 20px; }
.metric-title { font-size: 14px; color: #606266; margin-bottom: 12px; }
.metric-value { font-size: 32px; font-weight: bold; color: #303133; margin-bottom: 12px; }
.metric-value .unit { font-size: 16px; font-weight: normal; margin-left: 4px; }
.metric-footer { font-size: 13px; color: #909399; }
.text-danger { color: #F56C6C; }
.text-warning { color: #E6A23C; }
.font-bold { font-weight: bold; }

.charts-row { display: grid; grid-template-columns: repeat(2, 1fr); gap: 20px; margin-bottom: 24px; }
.chart-card { border-radius: 8px; border: none; }
.card-header { font-weight: bold; color: #303133; font-size: 15px; }
.chart-container { height: 300px; width: 100%; }

.bottom-row { display: grid; grid-template-columns: 2fr 1fr; gap: 20px; }
.ai-insight-card, .timeline-card { border-radius: 8px; border: none; }
.ai-header { display: flex; align-items: center; gap: 8px; color: #722ed1; }
.insight-content { display: flex; flex-direction: column; gap: 16px; }
.insight-alert :deep(.el-alert__title) { font-weight: bold; font-size: 14px; }
.insight-alert :deep(.el-alert__description) { margin-top: 6px; font-size: 13px; line-height: 1.5; }

.timeline-inner-card { background-color: #f8f9fa; border: none; border-radius: 6px; }
.timeline-inner-card :deep(.el-card__body) { padding: 12px 16px; }
.timeline-inner-card h4 { margin: 0 0 8px 0; font-size: 14px; color: #303133; }
.timeline-inner-card p { margin: 0; font-size: 13px; color: #606266; }
</style>
