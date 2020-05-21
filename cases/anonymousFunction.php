<?php

$func1 = function ($var1) {

};

$var1 = new Schema;

Schema::create('checklists', function (Blueprint $table) use ($func1, $var1) {
    $table->bigIncrements('id');
    $table->timestamps();
});

usort($var1, function($el) {

});
