<?php

class TestClass {
    
}

class TestClass1 extends TestClass implements TestInterface {

}

class TestClass2 implements TestInterface, TestInterface2 {

}

class TestClass3 extends TestClass implements TestInterface, TestInterface2 {}
