<template>
  <div class="dashboard-page">
    <!-- 顶部：项目标题与 AI 健康度状态 -->
    <div class="page-header">
      <div class="header-left">
        <h2>电商后台重构项目 (V2.0)</h2>
        <el-tag type="warning" effect="dark" size="large" class="health-tag">
          <el-icon><Warning /></el-icon> 进度健康度：关注
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
        <div class="metric-value">68<span class="unit">%</span></div>
        <div class="metric-footer">
          总任务: 124 <el-divider direction="vertical" /> 已完成: 84
        </div>
      </el-card>
      <el-card shadow="hover" class="metric-card">
        <div class="metric-title">测试覆盖率</div>
        <div class="metric-value">85<span class="unit">%</span></div>
        <div class="metric-footer">
          核心功能点已全部覆盖测试用例
        </div>
      </el-card>
      <el-card shadow="hover" class="metric-card">
        <div class="metric-title">活跃缺陷</div>
        <div class="metric-value text-danger">12<span class="unit">个</span></div>
        <div class="metric-footer">
          阻塞验收: <span class="text-danger font-bold">2</span> <el-divider direction="vertical" /> 严重: 3
        </div>
      </el-card>
      <el-card shadow="hover" class="metric-card">
        <div class="metric-title">变更冲击</div>
        <div class="metric-value text-warning">5<span class="unit">项</span></div>
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
            title="核心延期风险"
            type="error"
            description="任务【对接飞书 OAuth2.0 获取 UserInfo】已超期 2 天，且有 2 个严重缺陷未修复，直接阻塞了【V1.0 基础功能】的测试验收。"
            show-icon
            :closable="false"
            class="insight-alert"
          />
          <el-alert
            title="测试覆盖缺口"
            type="warning"
            description="功能点【自动创建访客账号】尚未关联任何测试用例，建议立即让 AI 生成补充用例。"
            show-icon
            :closable="false"
            class="insight-alert"
          />
          <el-alert
            title="下一步建议"
            type="success"
            description="1. 优先推进处理 BUG-101 和 BUG-102，解除验收阻塞。 2. 确认最新的【扫码登录】需求变更范围草稿。"
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
import { ref, onMounted, onUnmounted, markRaw } from 'vue';
import * as echarts from 'echarts';
import { Warning, MagicStick } from '@element-plus/icons-vue';

const taskChartRef = ref(null);
const bugChartRef = ref(null);

let taskChart = null;
let bugChart = null;

onMounted(() => {
  initTaskChart();
  initBugChart();
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
          { value: 15, name: '待开发' },
          { value: 30, name: '开发中' },
          { value: 20, name: '待测试/测试中' },
          { value: 10, name: '待验收' },
          { value: 84, name: '已验收/已上线' }
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
        data: [1, 3, 15, 8]
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