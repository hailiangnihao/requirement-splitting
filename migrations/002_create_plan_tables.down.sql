-- 注意：回滚时需要注意表的外键依赖关系，被依赖的表（如 feature_points）需要后删
DROP TABLE IF EXISTS acceptance_items;
DROP TABLE IF EXISTS test_cases;
DROP TABLE IF EXISTS dev_tasks;
DROP TABLE IF EXISTS feature_points;
DROP TABLE IF EXISTS milestones;
DROP TABLE IF EXISTS modules;