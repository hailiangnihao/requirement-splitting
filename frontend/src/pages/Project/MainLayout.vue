<template>
  <div class="layout-container">
    <!-- 侧边栏导航 -->
    <aside class="sidebar">
      <div class="logo">
        <h2>AI项目管理台</h2>
      </div>
      <nav class="nav-menu">
        <router-link to="/projects" class="nav-item">项目列表</router-link>
        
        <!-- 选中项目后显示的具体项目导航 -->
        <div v-if="currentProjectId" class="project-menus">
          <div class="menu-divider">当前项目</div>
          <router-link :to="`/project/${currentProjectId}/overview`" class="nav-item">项目总览</router-link>
          <router-link :to="`/project/${currentProjectId}/split`" class="nav-item">需求拆分</router-link>
          <router-link :to="`/project/${currentProjectId}/kanban`" class="nav-item">任务看板</router-link>
          <router-link :to="`/project/${currentProjectId}/test-acceptance`" class="nav-item">测试验收</router-link>
          <router-link :to="`/project/${currentProjectId}/bugs`" class="nav-item">缺陷管理</router-link>
          <router-link :to="`/project/${currentProjectId}/changes`" class="nav-item">需求变更</router-link>
          <router-link :to="`/project/${currentProjectId}/risks`" class="nav-item">风险与缺口</router-link>
        </div>
      </nav>
    </aside>

    <!-- 右侧内容区 -->
    <main class="main-content">
      <header class="header">
        <div class="breadcrumb">当前位置：{{ route.meta.title || '首页' }}</div>
        <div class="user-info">
          <span class="role-badge">项目经理</span>
          <span class="avatar">PM</span>
        </div>
      </header>
      <div class="page-container">
        <router-view />
      </div>
    </main>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import { useRoute } from 'vue-router';

const route = useRoute();
// 实际开发中可以通过路由参数或全局状态(Pinia)获取当前选中的项目ID
const currentProjectId = computed(() => route.params.id || 'demo-1');
</script>

<style scoped>
.layout-container {
  display: flex;
  height: 100vh;
  background-color: #f5f7fa;
}
.sidebar {
  width: 240px;
  background-color: #001529;
  color: white;
  display: flex;
  flex-direction: column;
}
.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid #1f2d3d;
}
.nav-menu { flex: 1; padding: 16px 0; }
.nav-item { display: block; padding: 12px 24px; color: #a6adb4; text-decoration: none; }
.nav-item:hover, .router-link-active { background-color: #1890ff; color: white; }
.menu-divider { padding: 16px 24px 8px; font-size: 12px; color: #6b7280; }
.main-content { flex: 1; display: flex; flex-direction: column; overflow: hidden; }
.header { height: 60px; background: white; display: flex; align-items: center; justify-content: space-between; padding: 0 24px; box-shadow: 0 1px 4px rgba(0,21,41,.08); z-index: 10; }
.page-container { flex: 1; padding: 24px; overflow-y: auto; }
.avatar { display: inline-flex; align-items: center; justify-content: center; width: 32px; height: 32px; border-radius: 50%; background: #1890ff; color: white; margin-left: 12px; }
.role-badge { font-size: 12px; background: #e6f7ff; color: #1890ff; padding: 2px 8px; border-radius: 12px; border: 1px solid #91d5ff; }
</style>