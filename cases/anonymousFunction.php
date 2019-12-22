<?php

$func1 = function ($var1) {

};

Schema::create('checklists', function (Blueprint $table) {
    $table->bigIncrements('id');
    $table->timestamps();
});
