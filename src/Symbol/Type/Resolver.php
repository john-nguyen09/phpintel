<?php
declare(strict_types=1);
namespace PhpIntel\Symbol\Type;

use Microsoft\PhpParser;
use Microsoft\PhpParser\Node;
use Microsoft\PhpParser\Node\Statement;
use PhpIntel\ParserHelper;

final class Resolver
{
    public static function resolveReservedWordToType($node)
    {
        if (!(($token = $node->children) instanceof PhpParser\Token)) {
            return null;
        }

        switch ($token->kind) {
            case PhpParser\TokenKind::TrueReservedWord:
            case PhpParser\TokenKind::FalseReservedWord:
                return 'bool';
            case PhpParser\TokenKind::NullReservedWord:
                return 'null';
        }
    }

    public static function resolveExpressionToType($expr)
    {
        if ($expr instanceof Node\ReservedWord) {
            return self::resolveReservedWordToType($expr);
        }

        if (
            (
                $expr instanceof Node\Expression\BinaryExpression &&
                (
                    $expr->operator->kind === PhpParser\TokenKind::DotToken ||
                    $expr->operator->kind === PhpParser\TokenKind::DotEqualsToken
                )
            ) ||
            $expr instanceof Node\StringLiteral ||
            (
                $expr instanceof Node\Expression\CastExpression &&
                $expr->castType->kind === PhpParser\TokenKind::StringCastToken
            )
        ) {
            return 'string';
        }

        // INTEGER EXPRESSIONS: resolve to Types\Integer
        //   [literal] 1
        //   [operator] <=>, &, ^, |
        //   TODO: Magic constants (__LINE__)
        if (
            // TODO: consider different Node types of float/int, also better property name (not "children")
            (
                $expr instanceof Node\NumericLiteral &&
                $expr->children->kind === PhpParser\TokenKind::IntegerLiteralToken
            ) ||
            $expr instanceof Node\Expression\BinaryExpression &&
            (
                ($operator = $expr->operator->kind) &&
                (
                    $operator === PhpParser\TokenKind::LessThanEqualsGreaterThanToken ||
                    $operator === PhpParser\TokenKind::AmpersandToken ||
                    $operator === PhpParser\TokenKind::CaretToken ||
                    $operator === PhpParser\TokenKind::BarToken ||
                    $operator === PhpParser\TokenKind::PlusToken ||
                    $operator === PhpParser\TokenKind::MinusToken ||
                    $operator === PhpParser\TokenKind::AsteriskToken ||
                    $operator === PhpParser\TokenKind::SlashToken ||
                    $operator === PhpParser\TokenKind::PlusPlusToken ||
                    $operator === PhpParser\TokenKind::MinusMinusToken ||
                    $operator === PhpParser\TokenKind::AsteriskAsteriskEqualsToken ||
                    $operator === PhpParser\TokenKind::SlashEqualsToken
                )
            )
        ) {
            return 'int';
        }
        
        // FLOAT EXPRESSIONS: resolve to Types\Float
        //   [literal] 1.5
        //   [operator] /
        //   [cast] (double)
        if (
            $expr instanceof Node\NumericLiteral &&
            $expr->children->kind === PhpParser\TokenKind::FloatingLiteralToken ||
            (
                $expr instanceof Node\Expression\CastExpression &&
                $expr->castType->kind === PhpParser\TokenKind::DoubleCastToken
            ) ||
            (
                $expr instanceof Node\Expression\BinaryExpression &&
                $expr->operator->kind === PhpParser\TokenKind::SlashToken
            )
        ) {
            return 'float';
        }

        // BOOLEAN EXPRESSIONS: resolve to Types\Boolean
        //   (bool) $expression
        //   !$expression
        //   empty($var)
        //   isset($var)
        //   >, >=, <, <=, &&, ||, AND, OR, XOR, ==, ===, !=, !==
        if (
            ParserHelper\isBooleanExpression($expr) ||
            (
                $expr instanceof Node\Expression\CastExpression &&
                $expr->castType->kind === PhpParser\TokenKind::BoolCastToken
            ) ||
            (
                $expr instanceof Node\Expression\UnaryOpExpression &&
                $expr->operator->kind === PhpParser\TokenKind::ExclamationToken
            ) ||
            $expr instanceof Node\Expression\EmptyIntrinsicExpression ||
            $expr instanceof Node\Expression\IssetIntrinsicExpression
        ) {
            return 'bool';
        }

        if (
            $expr instanceof Node\Expression\ObjectCreationExpression &&
            ($typeDesignator = $expr->classTypeDesignator) !== null &&
            $typeDesignator instanceof Node\QualifiedName
        ) {
            return $typeDesignator->getText();
        }

        return 'mixed';
    }
}