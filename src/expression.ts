import { PhpFile } from "./phpFile";
import { Position, Range } from "./meta";
import { TreeTraverser } from "./treeTraverser";
import { ParserUtils } from "./util/parser";
import * as Parser from "tree-sitter";
import { Type, TypeComposite } from "./typeResolver/type";
import { TypeResolver } from "./typeResolver";

export class Expression {
    private _type: string | null = null;

    private _name: Type | undefined = undefined;
    private _nameRange: Range | undefined = undefined;
    private _scope: TypeComposite = new TypeComposite();
    private _scopeRange: Range | undefined = undefined;

    constructor(phpFile: PhpFile, pos: Position) {
        const traverser = new TreeTraverser(phpFile.getTree().rootNode);
        traverser.setPosition(pos);

        // const traverserForDebugging = new TreeTraverser(traverser.node);

        // const parentType = ParserUtils.getType(traverserForDebugging.parent());
        // const parentOfParentType = ParserUtils.getType(traverserForDebugging.parent());

        const nodeType = ParserUtils.getType(traverser.node);

        // console.log({
        //     nodeType, parentType, parentOfParentType
        // });

        if (nodeType === 'name') {
            const parent = traverser.node;
            const parentType = ParserUtils.getType(traverser.parent());

            if (parentType === 'qualified_name') {
                let type = ParserUtils.getType(traverser.parent());

                if (type === 'object_creation_expression') {
                    this.onTypeDesignator(traverser.node);
                } else if (type === 'function_call_expression') {
                    this.onFunctionCall(traverser.node);
                } else if (type === 'scoped_call_expression') {
                    this.onScopedCall(traverser.node);
                } else {
                    this.onConstant(parent);
                }
            } else if (parentType === 'scoped_call_expression') {
                this.onScopedCall(traverser.node);
            } else if (parentType === 'class_constant_access_expression') {
                this.onClassConstantAccess(traverser.node);
            } else if (parentType === 'member_call_expression') {
                this.onMemberCall(traverser.node);
            } else if (parentType === 'member_access_expression') {
                this.onMemberAccess(traverser.node);
            } else {
                let type = ParserUtils.getType(traverser.parent());

                if (type === 'scoped_property_access_expression') {
                    this.onScopedProperty(traverser.node);
                }
            }
        } else if (nodeType === '->') {
            const parentType = ParserUtils.getType(traverser.parent());

            if (parentType === 'member_access_expression') {
                this.onMemberAccess(traverser.node);
            }
        }
    }

    get type(): string {
        if (this._type === null) {
            return 'unkown';
        }

        return this._type;
    }

    get name(): Type | undefined {
        return this._name;
    }

    get nameRange(): Range | undefined {
        return this._nameRange;
    }

    get scope(): TypeComposite {
        return this._scope;
    }

    get scopeRange(): Range | undefined {
        return this._scopeRange;
    }

    get isKnown(): boolean {
        return this.type !== 'unknown';
    }

    private onTypeDesignator(node: Parser.SyntaxNode) {
        const nameNode = node.firstNamedChild;

        if (nameNode !== null) {
            this._type = 'type_designator';
            this._name = TypeResolver.stringToType(nameNode.text);
        }

        this._nameRange = ParserUtils.getRange(node);
    }

    private onFunctionCall(node: Parser.SyntaxNode) {
        const nameNode = node.firstNamedChild;

        if (nameNode !== null) {
            this._type = 'function_call';
            this._name = TypeResolver.stringToType(nameNode.text);
        }

        this._nameRange = ParserUtils.getRange(node);
    }

    private onConstant(node: Parser.SyntaxNode) {
        this._type = 'constant';

        this._name = TypeResolver.stringToType(node.text);
        this._nameRange = ParserUtils.getRange(node);
    }

    private onScopedClass(node: Parser.SyntaxNode) {
        this._scope.push(TypeResolver.stringToType(node.text));
        this._scopeRange = ParserUtils.getRange(node);
    }

    private onScopedName(node: Parser.SyntaxNode) {
        this._name = TypeResolver.stringToType(node.text);
        this._nameRange = ParserUtils.getRange(node);
    }

    private onScopedCall(node: Parser.SyntaxNode) {
        let hasColonColon = false;

        this._type = 'scoped_call';

        for (const child of node.children) {
            if (child.type === '::') {
                hasColonColon = true;
                continue;
            }

            if (!hasColonColon && child.type === 'qualified_name') {
                this.onScopedClass(child);
                continue;
            }

            if (hasColonColon && child.type === 'name') {
                this.onScopedName(child);
                break;
            }
        }
    }

    private onScopedProperty(node: Parser.SyntaxNode) {
        let hasColonColon = false;

        this._type = 'scoped_property_access';

        for (const child of node.children) {
            if (child.type === '::') {
                hasColonColon = true;
                continue;
            }

            if (!hasColonColon && child.type === 'qualified_name') {
                this.onScopedClass(child);
                continue;
            }

            if (hasColonColon && child.type === 'variable_name') {
                this.onScopedName(child);
                break;
            }
        }
    }

    private onClassConstantAccess(node: Parser.SyntaxNode) {
        let hasColonColon = false;

        this._type = 'class_constant_access';

        for (const child of node.children) {
            if (child.type === '::') {
                hasColonColon = true;
                continue;
            }

            if (!hasColonColon && child.type === 'qualified_name') {
                this.onScopedClass(child);
                continue;
            }

            if (hasColonColon && child.type === 'name') {
                this.onScopedName(child);
                break;
            }
        }
    }

    private onDereferencable(node: Parser.SyntaxNode) {
        // Trying to solve as much as possible however this will need to be looked up later
        for (const child of node.children) {
            if (child.type === 'variable_name') {
                this.onScopedClass(child);
                continue;
            }
        }
    }

    private onMemberCall(node: Parser.SyntaxNode) {
        let hasArrow = false;

        this._type = 'member_call';

        for (const child of node.children) {
            if (child.type === '->') {
                hasArrow = true;
                continue;
            }

            if (!hasArrow && child.type === 'dereferencable_expression') {
                this.onDereferencable(child);
                continue;
            }

            if (hasArrow && child.type === 'name') {
                this.onScopedName(child);
                break;
            }
        }
    }

    private onMemberAccess(node: Parser.SyntaxNode) {
        let hasArrow = false;

        this._type = 'member_access';

        for (const child of node.children) {
            if (child.type === '->') {
                hasArrow = true;
                continue;
            }

            if (!hasArrow && child.type === 'dereferencable_expression') {
                this.onDereferencable(child);
                continue;
            }

            if (hasArrow && child.type === 'name') {
                this.onScopedName(child);
                break;
            }
        }
    }
}