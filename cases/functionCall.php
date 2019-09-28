<?php

$file = fopen(__FILE__, 'r');

while (($line = fgets($file))) {
    printf($line . PHP_EOL);
}

fclose($file);
