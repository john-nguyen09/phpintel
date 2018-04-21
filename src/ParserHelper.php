<?php
declare (strict_types = 1);
namespace PhpIntel\ParserHelper;

use Microsoft\PhpParser;
use Microsoft\PhpParser\Node;

/**
 * Returns true if the node is a usage of `define`.
 * e.g. define('TEST_DEFINE_CONSTANT', false);
 * @param Node $node
 * @return bool
 */
function isConstDefineExpression(Node $node) : bool
{
    return $node instanceof Node\Expression\CallExpression
        && $node->callableExpression instanceof Node\QualifiedName
        && strtolower($node->callableExpression->getText()) === 'define'
        && isset($node->argumentExpressionList->children[0])
        && $node->argumentExpressionList->children[0]->expression instanceof Node\StringLiteral
        && isset($node->argumentExpressionList->children[2]);
}

/**
 * Checks whether the given Node declares the given variable name
 *
 * @param Node $n The Node to check
 * @param string $name The name of the wanted variable
 * @return bool
 */
function isVariableDeclaration(Node $n, string $name) {
    if (
        // TODO - clean this up
        (
            $n instanceof Node\Expression\AssignmentExpression &&
            $n->operator->kind === PhpParser\TokenKind::EqualsToken
        ) &&
        $n->leftOperand instanceof Node\Expression\Variable &&
        $n->leftOperand->getName() === $name
    ) {
        return true;
    }

    if (
        (
            $n instanceof Node\ForeachValue ||
            $n instanceof Node\ForeachKey
        ) &&
        $n->expression instanceof Node\Expression\Variable &&
        $n->expression->getName() === $name
    ) {
        return true;
    }

    return false;
}

function isBooleanExpression($expression) : bool
{
    if (!($expression instanceof Node\Expression\BinaryExpression)) {
        return false;
    }
    switch ($expression->operator->kind) {
        case PhpParser\TokenKind::InstanceOfKeyword:
        case PhpParser\TokenKind::GreaterThanToken:
        case PhpParser\TokenKind::GreaterThanEqualsToken:
        case PhpParser\TokenKind::LessThanToken:
        case PhpParser\TokenKind::LessThanEqualsToken:
        case PhpParser\TokenKind::AndKeyword:
        case PhpParser\TokenKind::AmpersandAmpersandToken:
        case PhpParser\TokenKind::LessThanEqualsGreaterThanToken:
        case PhpParser\TokenKind::OrKeyword:
        case PhpParser\TokenKind::BarBarToken:
        case PhpParser\TokenKind::XorKeyword:
        case PhpParser\TokenKind::ExclamationEqualsEqualsToken:
        case PhpParser\TokenKind::ExclamationEqualsToken:
        case PhpParser\TokenKind::CaretToken:
        case PhpParser\TokenKind::EqualsEqualsEqualsToken:
        case PhpParser\TokenKind::EqualsToken:
            return true;
    }
    return false;
}