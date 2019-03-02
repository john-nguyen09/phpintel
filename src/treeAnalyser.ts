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
import { Trait } from "./trait";
import { Interface } from "./interface";
import { ClassConstant } from "./classConstant";
import { Property } from "./property";
import { ParserUtils } from "./util/parser";

export namespace TreeAnalyser {
    const SCOPE_CLASS_TYPES = new Map<string, boolean>([
        ['class_declaration', true],
        ['interface_declaration', true],
        ['trait_declaration', true],
    ]);

    const VISIBILITY_MODIFIER_TYPES = new Map<string, VisibilityModifier>([
        ['public', VisibilityModifier.Public],
        ['protected', VisibilityModifier.Protected],
        ['private', VisibilityModifier.Private],
    ]);

    export function analyse(phpFile: PhpFile) {
        const tree = phpFile.getTree();

        // const astString = Formatter.treeSitterOutput(tree.rootNode.toString());
        // const debugDir = path.join(__dirname, '..', 'debug');
        // fs.writeFile(path.join(debugDir, path.basename(phpFile.path) + '.ast'), astString, (err) => {
        //     if (err) {
        //         console.log(err);
        //     }
        // });

        traverse(phpFile, tree.rootNode);

        // console.log(inspect(phpFile, {depth: 7}));

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
            const theFunction = onFunction(phpFile, node);
            phpFile.pushFunction(theFunction);

            return false;
        }

        if (node.type === 'const_declaration') {
            const constants = onConstant(phpFile, node);
            constants.forEach((constant) => {
                phpFile.pushConstant(constant);
            });

            return false;
        }

        if (node.type === 'function_call_expression') {
            const firstChild = node.firstChild;

            if (firstChild !== null && firstChild.type == 'qualified_name' && firstChild.text == 'define') {
                const defineConstant = onDefineConstant(phpFile, node);
                phpFile.pushConstant(defineConstant);

                return false;
            }
        }

        if (node.type === 'class_declaration') {
            const theClass = onClass(phpFile, node);

            phpFile.pushClass(theClass);
            phpFile.pushScopeClass(theClass);

            return true; // There will symbols inside classes so keep descending
        }

        if (node.type === 'class_const_declaration') {
            const classConstants = onClassConstant(phpFile, node);
            const scopeClass = phpFile.scopeClass;

            if (scopeClass instanceof Class || scopeClass instanceof Interface) {
                classConstants.forEach((classConstant) => {
                    scopeClass.constants.push(classConstant);
                });
            }

            return false;
        }

        if (node.type === 'property_declaration') {
            const properties = onProperty(phpFile, node);
            const scopeClass = phpFile.scopeClass;

            if (scopeClass instanceof Class || scopeClass instanceof Trait) {
                properties.forEach((property) => {
                    scopeClass.properties.push(property);
                })
            }

            return false;
        }

        if (node.type === 'method_declaration') {
            const method = onMethod(phpFile, node);
            const scopeClass = phpFile.scopeClass;

            if (scopeClass !== undefined) {
                scopeClass.methods.push(method);
            }

            return false;
        }

        if (node.type === 'trait_declaration') {
            const trait = onTrait(phpFile, node);

            phpFile.pushTrait(trait);
            phpFile.pushScopeClass(trait);

            return true;
        }

        if (node.type === 'interface_declaration') {
            const theInterface = onInterface(phpFile, node);

            phpFile.pushInterface(theInterface);
            phpFile.pushScopeClass(theInterface);

            return true;
        }

        return true;
    }

    function getLocation(phpFile: PhpFile, node: Parser.SyntaxNode): Location {
        return {
            uri: phpFile.uri,
            range: ParserUtils.getRange(node),
        };
    }

    function onFunction(phpFile: PhpFile, node: Parser.SyntaxNode): Function {
        const nameNode = node.firstNamedChild;
        const theFunction = new Function();

        theFunction.location = getLocation(phpFile, node);

        if (nameNode !== null) {
            theFunction.name = nameNode.text;
        }

        return theFunction;
    }

    function onConstant(phpFile: PhpFile, node: Parser.SyntaxNode): Constant[] {
        const constants: Constant[] = [];

        for (const constElement of node.children) {
            if (constElement.type == 'const_element') {
                constants.push(onConstantElement(phpFile, constElement));
            }
        }

        return constants;
    }

    function onConstantElement(phpFile: PhpFile, node: Parser.SyntaxNode): Constant {
        const theConst = new Constant();
        let value: string = '';
        let hasEqual: boolean = false;

        theConst.location = getLocation(phpFile, node);

        for (const child of node.children) {
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

        return theConst;
    }

    function onDefineConstant(phpFile: PhpFile, node: Parser.SyntaxNode): DefineConstant {
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

        return defineConstant;
    }

    function onClass(phpFile: PhpFile, node: Parser.SyntaxNode): Class {
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

        return theClass;
    }

    function onTrait(phpFile: PhpFile, node: Parser.SyntaxNode): Trait {
        const nameNode = node.firstNamedChild;
        const trait = new Trait();

        trait.location = getLocation(phpFile, node);

        if (nameNode !== null) {
            trait.name = nameNode.text;
        }

        return trait;
    }

    function onInterface(phpFile: PhpFile, node: Parser.SyntaxNode): Interface {
        const nameNode = node.firstNamedChild;
        const theInterface = new Interface();

        theInterface.location = getLocation(phpFile, node);

        if (nameNode !== null) {
            theInterface.name = nameNode.text;
        }

        return theInterface;
    }

    function onMethod(phpFile: PhpFile, node: Parser.SyntaxNode): Method {
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

                method.extends(theFunction);
            }
        }

        const scopeClass = phpFile.scopeClass;
        if (scopeClass !== undefined) {
            method.scope = scopeClass.name;
        }

        return method;
    }

    function onClassConstant(phpFile: PhpFile, node: Parser.SyntaxNode): ClassConstant[] {
        const classConstants: ClassConstant[] = [];
        const scopeClass = phpFile.scopeClass;

        for (const child of node.children) {
            if (child.type == 'const_element') {
                const classConstant = new ClassConstant();

                classConstant.extends(onConstantElement(phpFile, child));

                if (scopeClass !== undefined) {
                    classConstant.scope = scopeClass.name;
                }

                classConstants.push(classConstant);
            }
        }

        return classConstants;
    }

    function onProperty(phpFile: PhpFile, node: Parser.SyntaxNode): Property[] {
        const properties: Property[] = [];
        let propertyVisibility = VisibilityModifier.Public;
        let isStatic: boolean = false;
        const scopeClass = phpFile.scopeClass;

        for (const child of node.children) {
            if (child.type === 'property_modifier') {
                for (const modifier of child.children) {
                    if (modifier.type === 'visibility_modifier') {
                        const visibility = VISIBILITY_MODIFIER_TYPES.get(child.text);

                        if (visibility !== undefined) {
                            propertyVisibility = visibility;
                        }
                    } else if (modifier.type === 'static_modifier') {
                        isStatic = true;
                    }
                }
            } else if (child.type === 'property_element') {
                const property = new Property();
                const nameNode = child.firstNamedChild;

                property.location = getLocation(phpFile, child);
                property.modifier = {
                    visibility: propertyVisibility,
                    static: isStatic,
                };

                if (nameNode !== null) {
                    property.name = nameNode.text;
                }

                if (scopeClass !== undefined) {
                    property.scope = scopeClass.name;
                }

                properties.push(property);
            }
        }

        return properties;
    }
}