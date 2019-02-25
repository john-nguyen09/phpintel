<?php

$DB->;

$OUTPUT->;

function global_variable_func1() {
    // No completion
    $DB->
}

function global_variable_func2() {
    global $DB;

    // Completion
    $DB->
}