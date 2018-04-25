<?php
declare(strict_types=1);
namespace PhpIntel\Symbol\Type;

use Microsoft\PhpParser;
use Microsoft\PhpParser\Node;
use Microsoft\PhpParser\Node\Statement;
use Microsoft\PhpParser\Expression;
use PhpIntel\NodeTraverser;

class FunctionResolver
{
    public static function resolveFunctionType(Phpparser\FunctionLike $node)
    {
        // TODO: docblock type

        return self::resolveReturnType($node);
    }

    public static function resolveReturnType(Phpparser\FunctionLike $node)
    {
        if (
            $node->compoundStatementOrSemicolon === null ||
            !($node->compoundStatementOrSemicolon instanceof Statement\CompoundStatementNode) ||
            \count($node->compoundStatementOrSemicolon->statements) === 0
        ) {
            return [];
        }

        $types = [];

        foreach ($node->compoundStatementOrSemicolon->statements as $stmt) {
            NodeTraverser::traverseChildren($stmt, function($node) use (&$types) {
                if ($node instanceof Statement\ReturnStatement) {
                    if ($node->expression === null) {
                        return;
                    }

                    $types[] = Resolver::resolveExpressionToType($node->expression);
                }
            });
        }

        return array_unique($types);
    }
}