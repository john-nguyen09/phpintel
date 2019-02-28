import { PhpFile } from "./phpFile";
import * as Parser from "tree-sitter";
import { Function } from "./function";
import { Constant } from "./constant";
import { Formatter } from "./util/formatter";
import * as path from "path";
import * as fs from "fs";
import { DefineConstant } from "./defineConstant";
import { Class } from "./class";
import { Location } from "./meta";
import { Method } from "./method";
import { StaticModifier, VisibilityModifier } from "./modifier";
import { inspect } from "util";

export namespace TreeAnalyser {
    const SCOPE_CLASS_TYPES = new Map<string, boolean>([
        ['class_declaration', true],
        // ['interface_declaration', true],
        // ['trait_declaration', true],
    ]);

    const VISIBILITY_MODIFIER_TYPES = new Map<string, VisibilityModifier>([
        ['public', VisibilityModifier.Public],
        ['protected', VisibilityModifier.Protected],
        ['private', VisibilityModifier.Private],
    ]);

    export function analyse(phpFile: PhpFile) {
        const tree = phpFile.parse();

        const astString = Formatter.treeSitterOutput(tree.rootNode.toString());
        const debugDir = path.join(__dirname, '..', 'debug');
        fs.writeFile(path.join(debugDir, path.basename(phpFile.path) + '.ast'), astString, (err) => {
            if (err) {
                console.log(err);
            }
        });

        traverse(phpFile, tree.rootNode);

        console.log(inspect(phpFile.classes, {depth: 5}));

        return phpFile;
    }

    function traverse(phpFile: PhpFile, node: Parser.SyntaxNode): void {
        const shouldDescend = collectDefinitions(phpFile, node);

        if (shouldDescend) {
            for (const child of node.children) {
                traverse(phpFile, child);
            }
        }

        if (SCOPE_CLASS_TYPES.has(node.type)) {
            phpFile.popScopeClass();
        }
    }

    function collectDefinitions(phpFile: PhpFile, node: Parser.SyntaxNode): boolean {
        if (node.type === 'function_definition') {
            onFunction(phpFile, node);

            return false;
        }

        if (node.type === 'const_declaration') {
            onConstant(phpFile, node);

            return false;
        }

        if (node.type === 'function_call_expression') {
            const firstChild = node.firstChild;

            if (firstChild !== null && firstChild.type == 'qualified_name' && firstChild.text == 'define') {
                onDefineConstant(phpFile, node);

                return false;
            }
        }

        if (node.type === 'class_declaration') {
            onClass(phpFile, node);

            return true; // There will symbols inside classes so keep descending
        }

        if (node.type === 'method_declaration') {
            onMethod(phpFile, node);

            return false;
        }

        return true;
    }

    function getLocation(phpFile: PhpFile, node: Parser.SyntaxNode): Location {
        return {
            uri: phpFile.uri,
            range: {
                start: node.startPosition,
                end: node.endPosition,
            }
        };
    }

    function onFunction(phpFile: PhpFile, node: Parser.SyntaxNode): Function {
        const nameNode = node.firstNamedChild;
        const theFunction = new Function();

        theFunction.location = getLocation(phpFile, node);

        if (nameNode !== null) {
            theFunction.name = nameNode.text;
        }

        phpFile.pushFunction(theFunction);

        return theFunction;
    }

    function onConstant(phpFile: PhpFile, node: Parser.SyntaxNode) {
        for (const constElement of node.children) {
            if (constElement.type == 'const_element') {
                const theConst = new Constant();
                let value: string = '';
                let hasEqual: boolean = false;

                theConst.location = getLocation(phpFile, constElement);

                for (const child of constElement.children) {

                    if (child.type == 'name') {
                        theConst.name = child.text;
                        continue;
                    }
                    if (child.type == '=') {
                        hasEqual = true;
                        continue;
                    }

                    if (hasEqual) {
                        value += child.text;
                    }
                }

                theConst.value = value;

                phpFile.pushConstant(theConst);
            }
        }
    }

    function onDefineConstant(phpFile: PhpFile, node: Parser.SyntaxNode) {
        const defineConstant = new DefineConstant();

        defineConstant.location = getLocation(phpFile, node);

        for (const child of node.children) {
            if (child.type == 'arguments') {
                let hasComma = false;

                for (const arg of child.children) {
                    if (arg.type === ',') {
                        hasComma = true;

                        continue;
                    }

                    if (arg.type === 'string' && !hasComma) {
                        const value = arg.text;
                        defineConstant.name = value.substr(1, value.length - 2);

                        continue;
                    }

                    if (hasComma && arg.type !== ')') {
                        defineConstant.value += arg.text;
                    }
                }
            }
        }

        phpFile.pushConstant(defineConstant);
    }

    function onClass(phpFile: PhpFile, node: Parser.SyntaxNode) {
        const nameNode = node.firstNamedChild;
        const theClass = new Class();

        theClass.location = getLocation(phpFile, node);

        if (nameNode !== null) {
            theClass.name = nameNode.text;
        }

        for (const child of node.children) {
            if (child.type == 'class_base_clause') {
                for (const baseClassNode of child.children) {
                    if (baseClassNode.type == 'qualified_name') {
                        theClass.extends.push(baseClassNode.text);
                    }
                }
            } else if (child.type == 'class_interface_clause') {
                for (const interfaceNode of child.children) {
                    if (interfaceNode.type == 'qualified_name') {
                        theClass.implements.push(interfaceNode.text);
                    }
                }
            }
        }

        phpFile.pushClass(theClass);
        phpFile.pushScopeClass(theClass);
    }

    function onMethod(phpFile: PhpFile, node: Parser.SyntaxNode) {
        const method = new Method();
        method.location = getLocation(phpFile, node);

        for (const child of node.children) {
            if (child.type === 'visibility_modifier') {
                const modifier = child.firstChild;

                if (modifier !== null) {
                    const visibility = VISIBILITY_MODIFIER_TYPES.get(modifier.type);

                    if (visibility !== undefined) {
                        method.modifier.visibility = visibility;
                    } else {
                        method.modifier.visibility = VisibilityModifier.Public; // public by default
                    }
                }
            } else if (child.type === 'static_modifier') {
                method.modifier.static = StaticModifier.Static;
            } else if (child.type === 'function_definition') {
                const theFunction = onFunction(phpFile, child);

                method.extendsFromFunction(theFunction);
            }
        }

        const scopeClass = phpFile.scopeClass;
        if (scopeClass !== undefined) {
            method.scope = scopeClass.name;
        }

        phpFile.pushMethod(method);
    }
}