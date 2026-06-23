pg_dump -c -U postgres -d project_manager -h localhost | gzip > project_manager_`date +%Y-%m-%d"_"%H_%M_%S`.sql.gz
