// 字段名中文映射工具
export const fieldLabels = {
  // 计划树主要结构
  project_id: '项目ID',
  modules: '模块',
  milestones: '里程碑',
  feature_points: '功能点',
  dev_tasks: '开发任务',
  test_cases: '测试用例',
  acceptance_items: '验收项',

  // 模块相关字段
  module_id: '模块ID',
  module_key: '模块键',
  name: '名称',
  title: '标题',
  description: '描述',

  // 功能点相关
  feature_point_id: '功能点ID',
  feature_point_key: '功能点键',

  // 任务相关
  task_id: '任务ID',
  key: '键',
  status: '状态',
  priority: '优先级',
  assignee: '负责人',

  // 测试相关
  test_case_id: '测试用例ID',
  expected_result: '预期结果',
  actual_result: '实际结果',

  // 验收相关
  pass_criteria: '通过标准',

  // 通用字段
  id: 'ID',
  created_at: '创建时间',
  updated_at: '更新时间',
  created_by: '创建人',
  type: '类型',
  action: '操作',
  reason: '原因'
};

// 将字段名转换为中文标签
export function getFieldLabel(fieldName) {
  return fieldLabels[fieldName] || fieldName;
}

// 批量转换字段名
export function mapFieldsToLabels(obj) {
  if (!obj || typeof obj !== 'object') return obj;

  const mapped = {};
  for (const [key, value] of Object.entries(obj)) {
    const label = getFieldLabel(key);
    mapped[label] = value;
  }
  return mapped;
}

// 为数据添加中文标签
export function addChineseLabels(data) {
  if (!data) return data;

  // 处理数组
  if (Array.isArray(data)) {
    return data.map(item => addChineseLabels(item));
  }

  // 处理对象
  if (typeof data === 'object') {
    const result = { ...data };

    // 为特定字段添加中文标签
    if (result.modules) {
      result['模块列表'] = result.modules;
    }
    if (result.milestones) {
      result['里程碑列表'] = result.milestones;
    }
    if (result.feature_points) {
      result['功能点列表'] = result.feature_points;
    }
    if (result.dev_tasks) {
      result['开发任务列表'] = result.dev_tasks;
    }
    if (result.test_cases) {
      result['测试用例列表'] = result.test_cases;
    }
    if (result.acceptance_items) {
      result['验收项列表'] = result.acceptance_items;
    }

    return result;
  }

  return data;
}
