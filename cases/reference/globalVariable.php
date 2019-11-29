<?php

$DB->;

$OUTPUT->;

$DB->testMethod3();
$OUTPUT->testMethod3();

function global1() {
    $DB->; // No types
}

function global2() {
    global $DB;
    $DB->testMethod3(); // Types
    $var1 = $DB;
}
