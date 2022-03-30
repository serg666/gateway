begin;
alter table transactions add column customer text;
update transactions set customer='1';
alter table transactions alter column customer set not null;
commit;
