pg_dump -c -U postgres -d project_manager_test -h localhost | gzip > project_manager_test`date +%Y-%m-%d"_"%H_%M_%S`.sql.gz
