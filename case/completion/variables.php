<?php

$var1 = true;

$

function func1() {
    $var2 = false;
    $var3 = 1;

    $
}

$v

function get_child_nodes_for_node($node_id) {
    global $DB;

    $child_nodes = [];

    $child_nodes_from_db = $DB->get_records('hierarchy_node', [
        'parent_node_id' => $node_id
    ]);
    $

    if (!empty($child_nodes_from_db)) {
        foreach ($child_nodes_from_db as $child_node) {
            $child_nodes[] = $child_node->id;

            $child_nodes = array_merge($child_nodes, get_child_nodes_for_node($child_node->id));
        }
    }

    return $child_nodes;
}
