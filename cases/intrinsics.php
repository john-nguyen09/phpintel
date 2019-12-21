<?php

$var1 = new DateTime();

if (isset($var1)) {

}

unset($var1);

if (!empty($var1)) {
    eval('<?php ' . $var1);
}
