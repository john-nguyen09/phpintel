<?php
require_once('../../../config.php');
require_once($CFG->libdir . '/adminlib.php');
include("lib.php");
require_once($CFG->dirroot . '/blocks/coach/lib.php');
if (!isset($userid)) $userid = $USER->id;
require_login(0, false);
$context_system = context_system::instance();
$PAGE->set_context($context_system);

$PAGE->set_pagelayout('admin');
$PAGE->set_title($SITE->fullname);
$PAGE->set_heading(get_string('assign:addnewinstructor', 'tool_instructor'));
$PAGE->set_url($CFG->wwwroot . '/admin/tool/instructor/assign_instructor_add.php');
admin_externalpage_setup('tool_expertise_assign');

require_capability('tool/instructor:manage_instructors', $context_system);

$id = "";
$html = "";
if (isset($_POST['sub'])) {
    $id = $_POST['id'];
    if ($id == "") $id = $_POST['instructor'];
	// echo "<pre>".print_r($_POST,TRUE)."</pre>";
	// die();
    if (isset($_POST['areas'])) {
        $update = $_POST['areas'];
        $old = $DB->get_records('instructor_assign_expertise', array('userid' => $id));
        $existing = array();
        if (!empty($old)) {
            foreach ($old as $row) {
                if (!in_array($row->expertiseid, $update)) {
                    $DB->delete_records('instructor_assign_expertise', array('userid' => $id, 'expertiseid' => $row->expertiseid));
                } else {
                    $existing[] = $row->expertiseid;
                }
            }
        }
        if (!empty($update)) {
            foreach ($update as $key => $val) {
                if (in_array($val, $existing)) continue;
                $new = new stdClass();
                $new->expertiseid = $val;
                $new->userid = $id;
                $new->timemodified = time();
                $DB->insert_record('instructor_assign_expertise', $new);
            }
        }
        update_profile_field_value('IsInstructor', $id, MDL_CHECKBOX_CHECK);
        echo "<script type='text/javascript'>window.location='assign_instructor.php?u=1'</script>";
    } else $html .= get_string('expertise:errorempty', 'tool_instructor');
}

$add = "Add new";

$lock = "";
if (isset($_GET['id'])) {
    $id = $_GET['id'];
    $add = "Edit";
    $lock = "readOnly";
}
$instructors_options = get_instructors_options($id);
$area_options = get_expertise_area_options($id);

require_capability('tool/instructor:manage_instructors', $context_system);

$html_lib = "<script src='js/jquery-1.12.2.min.js'></script>
  <script src='js/jquery.tablesorter.js'></script>
  <script src='resource/chosen.jquery.js' type='text/javascript'></script>
  <link rel='stylesheet' href='resource/chosen.css'>
  <link rel='stylesheet' href='js/style.css'>";

$html .= get_string('assign:addnewinstructor:title', 'tool_instructor', $add);

$html .= "<form action='assign_instructor_add.php' method='POST'>";
$html .= "<table class='tablestyle10'>";
$html .= "<tr>";
$html .= "<td>" . get_string('assign:name', 'tool_instructor') . "</td>";
$html .= "<td><select name='instructor' class='chosen-select' $lock>" . $instructors_options . "</select></td>";
$html .= "</tr>";
$html .= "<tr>";
$html .= "<td>" . get_string('assign:expertise', 'tool_instructor') . "</td>";
$html .= "<td><select name='areas[]' class='chosen-select' multiple data-placeholder='Enter name'>" . $area_options . "</select></td>";
$html .= "</tr>";
$html .= "<tr>";
$html .= "<tr><td colspan='2'>";
$html .= "<input type='submit' name='sub' value='Save changes' style='margin-bottom: 0px;'>";
$html .= " <a href='assign_instructor.php' class='btn'> Cancel </a>";

$html .= "<input type='hidden' name='id' value='" . $id . "'>";

$html .= "</td></tr>";
$html .= "</table>";
$html .= "</form>";
$html .= "<br><br>";

echo $OUTPUT->header();
echo $html;
echo $OUTPUT->footer();
echo $html_lib;
?>
<script type="text/javascript">
var config = {
    '.chosen-select'           : {},
    '.chosen-select-deselect'  : {allow_single_deselect:true},
    '.chosen-select-no-single' : {disable_search_threshold:10},
    '.chosen-select-no-results': {no_results_text:'Could not find any!'},
    '.chosen-select-width'     : {width:"95%"}
}
for (var selector in config) {
    $(selector).chosen(config[selector]);
}
</script>