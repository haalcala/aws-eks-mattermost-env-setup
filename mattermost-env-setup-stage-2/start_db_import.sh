#!/bin/bash 

LOCAL_DB_HOST=$1
shift
LOCAL_DB_PORT=$1
shift
LOCAL_DB_NAME=$1
shift
LOCAL_DB_USER=$1
shift
LOCAL_DB_PASS=$1

shift
DUMP_FOLDER=$1

shift
DEST_DB_HOST=$1
shift
DEST_DB_PORT=$1
shift
DEST_DB_NAME=$1
shift
DEST_DB_USER=$1
shift
DEST_DB_PASS=$1

echo \$LOCAL_DB_HOST $LOCAL_DB_HOST
echo \$LOCAL_DB_PORT $LOCAL_DB_PORT
echo \$LOCAL_DB_NAME $LOCAL_DB_NAME
echo \$LOCAL_DB_USER $LOCAL_DB_USER
echo \$LOCAL_DB_PASS $LOCAL_DB_PASS
echo \$DUMP_FOLDER $DUMP_FOLDER
echo \$DEST_DB_HOST $DEST_DB_HOST
echo \$DEST_DB_PORT $DEST_DB_PORT
echo \$DEST_DB_NAME $DEST_DB_NAME
echo \$DEST_DB_USER $DEST_DB_USER
echo \$DEST_DB_PASS $DEST_DB_PASS

TABLES=$(mysql -h$LOCAL_DB_HOST -P$LOCAL_DB_PORT -u$LOCAL_DB_USER -p$LOCAL_DB_PASS $LOCAL_DB_NAME -e 'show tables' | grep -v Tables_in_$LOCAL_DB_NAME | awk '{print $1}')
# TABLES=Posts

echo \$TABLES $TABLES

SCHEMA_FILE=$DUMP_FOLDER/${LOCAL_DB_NAME}-schema.sql
SCHEMA_FILE_MYISAM=$DUMP_FOLDER/${LOCAL_DB_NAME}-schema-myisam.sql

echo \$SCHEMA_FILE $SCHEMA_FILE
echo \$SCHEMA_FILE_MYISAM $SCHEMA_FILE_MYISAM

timeit() {
	_start=`date '+%s'` ; $@ ; _end=`date '+%s'` ; elapsed=$(expr $_end - $_start)

	echo "$@ <<-- elapsed time $elapsed"
}

_start=`date '+%s'`
mysqldump --no-data -h$LOCAL_DB_HOST -P$LOCAL_DB_PORT -u$LOCAL_DB_USER -p$LOCAL_DB_PASS $LOCAL_DB_NAME $table > $SCHEMA_FILE
_end=`date '+%s'`
echo "mysqldump --no-data -h$LOCAL_DB_HOST -P$LOCAL_DB_PORT -u$LOCAL_DB_USER -p$LOCAL_DB_PASS $LOCAL_DB_NAME $table \> $SCHEMA_FILE <<-- elapsed time $(expr $_end - $_start)"

cp $SCHEMA_FILE  $SCHEMA_FILE_MYISAM

for table in $TABLES; do
	echo "----==== Dumping Schema ====----"
	DUMP_FILE=$DUMP_FOLDER/${LOCAL_DB_NAME}-${table}.sql

	echo \$DUMP_FILE $DUMP_FILE

	_start=`date '+%s'`
	mysqldump --skip-lock-tables --skip-add-locks --skip-add-drop-table --no-create-info -h$LOCAL_DB_HOST -P$LOCAL_DB_PORT -u$LOCAL_DB_USER -p$LOCAL_DB_PASS $LOCAL_DB_NAME $table > $DUMP_FILE
	_end=`date '+%s'`
	echo "----==== Dumping table: $table to $DUMP_FILE ..."
	echo "----==== Dumping table: $table to $DUMP_FILE <<-- elapsed time $(expr $_end - $_start)"

	# sed -e "s/^CREATE TABLE \`\($table\)\`.*/CREATE TABLE \`\1_I\`/g" $SCHEMA_FILE >  $SCHEMA_FILE_MYISAM.tmp
	sed -e "s/\`\($table\)\`/\`\1_I\`/g" $SCHEMA_FILE_MYISAM >  $SCHEMA_FILE_MYISAM.tmp
	sed -e "s/^\() ENGINE=\)\(InnoDB\)\(.*\)/\1MyISAM\3/g" $SCHEMA_FILE_MYISAM.tmp >  $SCHEMA_FILE_MYISAM

	rm -f $SCHEMA_FILE_MYISAM.tmp

	sed -e "s/^\(INSERT INTO \`\)\($table\)\(\` VALUES.*\)/\1\2_I\3/g" $DUMP_FILE >  $DUMP_FILE.tmp
	rm -f $DUMP_FILE
	mv -f $DUMP_FILE.tmp $DUMP_FILE

	sed -e 's/^\/\*!40000 ALTER TABLE \`.*\` ENABLE KEYS \*\/;//g' $DUMP_FILE >  $DUMP_FILE.tmp
	rm -f $DUMP_FILE
	mv -f $DUMP_FILE.tmp $DUMP_FILE
done

_start=`date '+%s'`
mysql -h$DEST_DB_HOST -P$DEST_DB_PORT -u$DEST_DB_USER -p$DEST_DB_PASS $DEST_DB_NAME < $SCHEMA_FILE
_end=`date '+%s'`
echo "mysql -h$DEST_DB_HOST -P$DEST_DB_PORT -u$DEST_DB_USER -p$DEST_DB_PASS $DEST_DB_NAME \< $SCHEMA_FILE <<-- elapsed time $(expr $_end - $_start)"
_start=`date '+%s'`
mysql -h$DEST_DB_HOST -P$DEST_DB_PORT -u$DEST_DB_USER -p$DEST_DB_PASS $DEST_DB_NAME < $SCHEMA_FILE_MYISAM
_end=`date '+%s'`
echo "mysql -h$DEST_DB_HOST -P$DEST_DB_PORT -u$DEST_DB_USER -p$DEST_DB_PASS $DEST_DB_NAME \< $SCHEMA_FILE_MYISAM <<-- elapsed time $(expr $_end - $_start)"

for table in $TABLES; do
	echo
	echo "---------------------------------------------------------------------------------"
	DUMP_FILE=$DUMP_FOLDER/${LOCAL_DB_NAME}-${table}.sql

	_start=`date '+%s'`
	mysql -h$DEST_DB_HOST -P$DEST_DB_PORT -u$DEST_DB_USER -p$DEST_DB_PASS $DEST_DB_NAME < $DUMP_FILE
	_end=`date '+%s'`
	echo "----==== Restoring $DUMP_FILE ..."
	echo "----==== Restoring $DUMP_FILE <<-- elapsed time $(expr $_end - $_start)"

	_start=`date '+%s'`
	mysql -h$DEST_DB_HOST -P$DEST_DB_PORT -u$DEST_DB_USER -p$DEST_DB_PASS $DEST_DB_NAME -e "delete from $table"
	_end=`date '+%s'`
	echo "----==== Deleting table: $table ..."
	echo "----==== Deleting table: $table <<-- elapsed time $(expr $_end - $_start)"
	_start=`date '+%s'`
	mysql -h$DEST_DB_HOST -P$DEST_DB_PORT -u$DEST_DB_USER -p$DEST_DB_PASS $DEST_DB_NAME -e "insert into $table select * from ${table}_I"
	_end=`date '+%s'`
	echo "----==== insert into $table select \* from ${table}_I ..."
	echo "----==== insert into $table select \* from ${table}_I" <<-- elapsed time $(expr $_end - $_start)"
	_start=`date '+%s'`
	mysql -h$DEST_DB_HOST -P$DEST_DB_PORT -u$DEST_DB_USER -p$DEST_DB_PASS $DEST_DB_NAME -e "drop table ${table}_I"
	_end=`date '+%s'`
	echo "----==== drop table ${table}_I <<-- elapsed time $(expr $_end - $_start)"
done
