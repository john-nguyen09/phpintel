<?php

class moodle_database {
    /**
     * Returns the sql generator used for db manipulation.
     * Used mostly in upgrade.php scripts.
     * @return database_manager The instance used to perform ddl operations.
     * @see lib/ddl/database_manager.php
     */
    public function get_manager() {
    }
}

class database_manager() {
    public function table_exists() {

    }
}

$db = new moodle_database();
$db_man = $db->get_manager();

$db_man->table_exists();
