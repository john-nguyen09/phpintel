<?php

/**
 * @global ClassWithMethod $DB This is a global variable
 */
global $DB;

/**
 * @global object This variable has optional name and
 * will be assigned extra type in the same document
 */
global $OUTPUT;

$OUTPUT = new ClassWithMethod();
