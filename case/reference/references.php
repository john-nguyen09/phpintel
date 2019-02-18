<?php
$class1 = new TestClass();

test_function1();

$content = test_function();

$html .= "<td>" . test_function2('assign:expertise', 'tool_instructor') . "</td>";

$classInstance = new ClassWithMethod();

$var1 = ClassWithMethod::staticMethod(1, 2);
$var2 = ClassWithMethod::$staticVariable;
$var3 = ClassWithConst::IS_ACTIVE;
$var4 = TEST_DEFINE_CONSTANT;

$refVar = true;

if ($something) {
    $refVar = 2;
}

$refVar2 = $refVar;

$class1->testProperty;
$class1->testMethod();
$class1->$staticTestProperty;
$class1->TEST_CLASS_CONST;
