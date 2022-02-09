create table sp_alert_event
(
    id                 varchar(32)                         not null comment '主键' primary key,
    name               varchar(255)                        null comment '事件名称（通过ticket模板计算出来）',
    org_id             bigint                              null comment '机构ID',
    alert_group_id     varchar(500)                        null comment '告警分组id',
    alert_id           bigint                              null comment '关联的告警策略ID',
    alert_name         varchar(255)                        null comment '关联告警策略名称',
    alert_type         varchar(128)                        null comment '告警类型',
    alert_index        varchar(128)                        null comment '告警规则index',
    alert_level        varchar(20)                         null comment '告警级别',
    alert_source       varchar(20)                         null comment '告警来源',
    alert_subject      varchar(500)                        null comment '告警对象',
    alert_state        varchar(20)                         null comment '告警状态',
    rule_id            bigint                              null comment '告警规则id',
    rule_name          varchar(255)                        null comment '告警规则名称',
    expression_id      bigint                              null comment '关联的告警表达式ID',
    last_trigger_time  datetime                            null comment '最后触发时间',
    first_trigger_time datetime                            null comment '首次触发时间',
    alert_group        varchar(500)                        null comment '告警分组',
    scope              varchar(24)                         null comment '域',
    scope_id           varchar(64)                         null comment '域id',
    created            datetime  default CURRENT_TIMESTAMP not null comment '创建时间',
    updated            timestamp default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '更新时间',
    constraint sp_alert_event_alert_group_id_uindex
        unique (alert_group_id)
) comment '告警事件';

create table sp_alert_event_suppress
(
    id             varchar(64)                        not null comment '主键' primary key,
    alert_event_id varchar(64)                        not null comment '关联的告警事件id',
    suppress_type  varchar(20)                        null comment '抑制类型',
    expire_time    datetime                           null comment '失效时间',
    enable         bit                                null comment '是否启用',
    created        datetime default CURRENT_TIMESTAMP null comment '创建时间',
    updated        datetime default CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP null comment '更新时间',
    constraint sp_alert_event_suppress_alert_event_id_uindex
        unique (alert_event_id)
) comment '告警事件抑制设置';