import { TreeNode, isToken, isPhrase } from "./util/parseTree";
import { Phrase } from "php7parser";
import { Symbol, TokenSymbol, isConsumer, isTransform, isCollection, isDocBlockConsumer } from "./symbol/symbol";
import { PhpDocument } from "./symbol/phpDocument";
import { Class } from "./symbol/class/class";
import { ClassHeader } from "./symbol/class/header";
import { ClassExtend } from "./symbol/class/extend";
import { ClassImplement } from "./symbol/class/implement";
import { ClassTraitUse } from "./symbol/class/traitUse";
import { QualifiedNameList } from "./symbol/list/qualifiedNameList";
import { QualifiedName } from "./symbol/name/qualifiedName";
import { NamespaceName } from "./symbol/name/namespaceName";
import { FunctionCall } from "./symbol/function/functionCall";
import { ArgumentExpressionList } from "./symbol/argumentExpressionList";
import { ConstantAccess } from "./symbol/constant/constantAccess";
import { AdditiveExpression } from "./symbol/type/additiveExpression";
import { Constant } from "./symbol/constant/constant";
import { Identifier } from "./symbol/identifier";
import { ClassConstant } from "./symbol/constant/classConstant";
import { Function } from "./symbol/function/function";
import { Return } from "./symbol/type/return";
import { FunctionHeader } from "./symbol/function/functionHeader";
import { Parameter } from "./symbol/variable/parameter";
import { TypeDeclaration } from "./symbol/type/typeDeclaration";
import { SimpleVariable } from "./symbol/variable/simpleVariable";
import { ClassTypeDesignator } from "./symbol/class/typeDesignator";
import { Method } from "./symbol/function/method";
import { MethodHeader } from "./symbol/function/methodHeader";
import { Property } from "./symbol/variable/property";
import { PropertyInitialiser } from "./symbol/variable/propertyInitialiser";
import { MemberModifierList } from "./symbol/class/memberModifierList";
import { PropertyDeclaration } from "./symbol/variable/propertyDeclaration";
import { DocBlock } from "./symbol/docBlock";
import { NamespaceDefinition } from "./symbol/namespace/definition";
import { NamespaceUse } from "./symbol/namespace/Use";
import { NamespaceUseClause } from "./symbol/namespace/useClause";
import { NamespaceAliasClause } from "./symbol/namespace/aliasClause";
import { PhraseKind, TokenKind } from "./util/parser";
import { VariableAssignment } from "./symbol/variable/varibleAssignment";

export class SymbolParser {
    protected symbolStack: (Symbol | null)[] = [];
    protected doc: PhpDocument;
    protected lastDocBlock: DocBlock | null = null;

    constructor(doc: PhpDocument) {
        this.doc = doc;
        this.pushSymbol(this.doc);
    }

    traverse(tree: Phrase) {
        let depth = 0;

        this.realTraverse(tree, depth);
    }

    private realTraverse(node: TreeNode, depth: number) {
        this.preorder(node, depth);

        if ('children' in node) {
            for (let child of node.children) {
                this.realTraverse(child, depth + 1);
            }
        }

        this.postorder(node, depth);
    }

    public getTree(): PhpDocument {
        return this.doc;
    }

    getParentSymbol(): Symbol | null {
        return this.symbolStack[this.symbolStack.length - 1];
    }

    pushSymbol(symbol: Symbol | null) {
        this.symbolStack.push(symbol);
    }

    preorder(node: TreeNode, depth: number) {
        let parentSymbol = this.getParentSymbol();

        if (isToken(node)) {
            let tokenType: number = <number>node.tokenType;

            if (tokenType == TokenKind.DocumentComment) {
                this.lastDocBlock = new DocBlock(node, this.doc, depth);
            } else {
                let symbol = new TokenSymbol(node, this.doc);

                if (parentSymbol && isConsumer(parentSymbol)) {
                    parentSymbol.consume(symbol);
                }
            }
        } else if (isPhrase(node)) {
            const p = <Phrase>node;
            const phraseType: number = <number>p.phraseType;

            switch (phraseType) {
                case PhraseKind.NamespaceDefinition:
                    this.pushSymbol(new NamespaceDefinition(node, this.doc));
                    break;
                case PhraseKind.NamespaceName:
                    this.pushSymbol(new NamespaceName(node, this.doc));
                    break;
                case PhraseKind.QualifiedName:
                    this.pushSymbol(new QualifiedName(node, this.doc));
                    break;
                case PhraseKind.QualifiedNameList:
                    this.pushSymbol(new QualifiedNameList(p, this.doc));
                    break;
                case PhraseKind.Identifier:
                    this.pushSymbol(new Identifier(p, this.doc));
                    break;

                case PhraseKind.NamespaceUseDeclaration:
                    this.pushSymbol(new NamespaceUse(p, this.doc));
                    break;
                case PhraseKind.NamespaceUseClause:
                    this.pushSymbol(new NamespaceUseClause(p, this.doc));
                    break;
                case PhraseKind.NamespaceUseGroupClause:
                    this.pushSymbol(new NamespaceUseClause(p, this.doc));
                    break;
                case PhraseKind.NamespaceAliasingClause:
                    this.pushSymbol(new NamespaceAliasClause(p, this.doc));
                    break;

                case PhraseKind.ClassDeclaration:
                    this.pushSymbol(new Class(p, this.doc));
                    break;
                case PhraseKind.ClassDeclarationHeader:
                    this.pushSymbol(new ClassHeader(p, this.doc));
                    break;
                case PhraseKind.ClassBaseClause:
                    this.pushSymbol(new ClassExtend(p, this.doc));
                    break;
                case PhraseKind.ClassInterfaceClause:
                    this.pushSymbol(new ClassImplement(p, this.doc));
                    break;
                case PhraseKind.TraitUseClause:
                    this.pushSymbol(new ClassTraitUse(p, this.doc));
                    break;

                case PhraseKind.ConstElement:
                    this.pushSymbol(new Constant(p, this.doc));
                    break;
                case PhraseKind.FunctionCallExpression:
                    this.pushSymbol(new FunctionCall(p, this.doc));
                    break;
                case PhraseKind.ClassConstElement:
                    this.pushSymbol(new ClassConstant(p, this.doc));
                    break;
                case PhraseKind.ArgumentExpressionList:
                    this.pushSymbol(new ArgumentExpressionList(p, this.doc));
                    break;
                case PhraseKind.ConstantAccessExpression:
                    this.pushSymbol(new ConstantAccess(p, this.doc));
                    break;
                case PhraseKind.AdditiveExpression:
                    this.pushSymbol(new AdditiveExpression(p, this.doc));
                    break;

                case PhraseKind.FunctionDeclaration:
                    this.pushSymbol(new Function(p, this.doc));
                    break;
                case PhraseKind.FunctionDeclarationHeader:
                    this.pushSymbol(new FunctionHeader(p, this.doc));
                    break;
                case PhraseKind.MethodDeclaration:
                    this.pushSymbol(new Method(p, this.doc));
                    break;
                case PhraseKind.MethodDeclarationHeader:
                    this.pushSymbol(new MethodHeader(p, this.doc));
                    break;
                case PhraseKind.ParameterDeclaration:
                    this.pushSymbol(new Parameter(p, this.doc));
                    break;
                case PhraseKind.TypeDeclaration:
                    this.pushSymbol(new TypeDeclaration(p, this.doc));
                    break;
                case PhraseKind.ReturnStatement:
                    this.pushSymbol(new Return(p, this.doc));
                    break;
                case PhraseKind.SimpleAssignmentExpression:
                    this.pushSymbol(new VariableAssignment(p, this.doc));
                    break;
                case PhraseKind.SimpleVariable:
                    this.pushSymbol(new SimpleVariable(p, this.doc));
                    break;
                case PhraseKind.ClassTypeDesignator:
                    this.pushSymbol(new ClassTypeDesignator(p, this.doc));

                case PhraseKind.PropertyElement:
                    this.pushSymbol(new Property(p, this.doc));
                    break;
                case PhraseKind.PropertyInitialiser:
                    this.pushSymbol(new PropertyInitialiser(p, this.doc));
                    break;
                case PhraseKind.MemberModifierList:
                    this.pushSymbol(new MemberModifierList(p, this.doc));
                    break;
                case PhraseKind.PropertyDeclaration:
                    this.pushSymbol(new PropertyDeclaration(p, this.doc));
                    break;

                default:
                    this.pushSymbol(null);
            }

            let symbol = this.symbolStack[this.symbolStack.length - 1];

            if (symbol != null && isDocBlockConsumer(symbol) && this.lastDocBlock != null) {
                symbol.consumeDocBlock(this.lastDocBlock);

                this.lastDocBlock = null;
            }
        }
    }

    postorder(node: TreeNode, depth: number) {
        if (isToken(node)) {
            return;
        }

        let symbol = this.symbolStack.pop();

        if (symbol && isTransform(symbol) && symbol.realSymbol) {
            symbol = symbol.realSymbol;
        }

        if (symbol == null || symbol == undefined) {
            return;
        }

        for (let i = this.symbolStack.length - 1; i >= 0; i--) {
            let prev = this.symbolStack[i];

            if (!prev || !isConsumer(prev)) {
                continue;
            }

            if (isCollection(symbol)) {
                let isConsumed = false;

                for (let realSymbol of symbol.realSymbols) {
                    if (!realSymbol) {
                        continue;
                    }

                    isConsumed = prev.consume(realSymbol) || isConsumed;
                }

                if (isConsumed) {
                    break;
                }
            } else {
                if (prev.consume(symbol)) {
                    break;
                }
            }
        }
    }
}
