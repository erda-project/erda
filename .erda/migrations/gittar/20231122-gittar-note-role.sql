alter table `dice_repo_notes` add role varchar(32) default 'USER' not null comment '创建者的角色，例如：AI, User。默认为 User。' after `score`;
