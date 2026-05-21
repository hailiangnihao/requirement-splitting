import { createRouter, createWebHistory } from 'vue-router';
import MainLayout from '../pages/Project/MainLayout.vue';

const routes = [
  {
    path: '/',
    component: MainLayout,
    redirect: '/projects',
    children: [
      {
        path: 'projects',
        name: 'ProjectList',
        component: () => import('../pages/Project/ProjectList.vue'),
        meta: { title: '项目列表' }
      },
      {
        path: 'project/:id',
        redirect: to => `/project/${to.params.id}/overview`,
        children: [
          { path: 'overview', name: 'ProjectOverview', component: () => import('../pages/Project/ProjectOverview.vue'), meta: { title: '项目总览' } },
          { path: 'split', name: 'RequirementSplit', component: () => import('../pages/Project/RequirementSplit.vue'), meta: { title: '需求拆分页' } },
          { path: 'kanban', name: 'TaskKanban', component: () => import('./TaskKanban.vue'), meta: { title: '任务看板' } },
          { path: 'test-acceptance', name: 'TestAcceptance', component: () => import('./TestAcceptance.vue'), meta: { title: '测试验收页' } },
          { path: 'bugs', name: 'BugList', component: () => import('./BugList.vue'), meta: { title: '缺陷管理页' } },
          { path: 'changes', name: 'RequirementChanges', component: () => import('../pages/Project/RequirementChanges.vue'), meta: { title: '需求变更页' } },
          { path: 'risks', name: 'RisksAndGaps', component: () => import('../pages/Project/RisksAndGaps.vue'), meta: { title: '风险与缺口页' } }
        ]
      }
    ]
  }
];

const router = createRouter({
  history: createWebHistory(),
  routes
});

export default router;
