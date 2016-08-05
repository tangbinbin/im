create database if not exists im;
use im;

create table user_account(
	id bigint not null auto_increment,
	username varchar(40) not null default '',
	password varchar(100) not null default '',
	phone varchar(40) not null default '',
	token varchar(50) not null default '',
	created_time int not null default 0,
	primary key(id),
	unique key idx_username(username)
)engine=innodb default charset=utf8;
