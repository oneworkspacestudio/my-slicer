type Setting={name:string;label:string;unit?:string;min?:number;max?:number;step?:number;select?:boolean;options?:string[]};
const n=(key:string,label:string,unit:string,min:number,max:number,step=.01):Setting=>({name:'process_'+key,label,unit,min,max,step});
const e=(key:string,label:string,options:string[]):Setting=>({name:'process_'+key,label,select:true,options});
const group=(title:string,fields:Setting[])=>({title,fields});
export function processGroups(tab:string,advanced:boolean){
 switch(tab){
 case 'Quality':return [group('First layer',[n('initial_layer_print_height','Initial layer height','mm',.08,.4)]),...(advanced?[group('Line width',[
  n('line_width','Default','mm',.2,1.2),n('initial_layer_line_width','Initial layer','mm',.2,1.2),n('outer_wall_line_width','Outer wall','mm',.2,1.2),n('inner_wall_line_width','Inner wall','mm',.2,1.2),n('top_surface_line_width','Top surface','mm',.2,1.2),n('sparse_infill_line_width','Sparse infill','mm',.2,1.2),n('internal_solid_infill_line_width','Internal solid infill','mm',.2,1.2),n('support_line_width','Support','mm',.2,1.2)])]:[]),group('Seam',[e('seam_position','Seam position',['nearest','aligned','back','random'])])];
 case 'Strength':return [group('Top / bottom shells',[n('top_shell_layers','Top shell layers','layers',0,50,1),n('bottom_shell_layers','Bottom shell layers','layers',0,50,1)])];
 case 'Speed':return [group('Initial layer',[n('initial_layer_speed','Initial layer','mm/s',1,100,1)]),group('Other layers',[n('outer_wall_speed','Outer wall','mm/s',1,500,1),n('inner_wall_speed','Inner wall','mm/s',1,500,1),n('sparse_infill_speed','Sparse infill','mm/s',1,500,1),n('top_surface_speed','Top surface','mm/s',1,500,1),n('travel_speed','Travel','mm/s',1,500,1)])];
 case 'Support':return [group('Type & style',[e('support_type','Type',['normal(auto)','tree(auto)','normal(manual)','tree(manual)']),e('support_style','Style',['default','grid','snug','tree_slim','tree_strong','tree_hybrid','tree_organic']),n('support_threshold_angle','Threshold angle','°',0,90,1)]),group('Raft',[n('raft_layers','Raft layers','layers',0,20,1)]),...(advanced?[group('Advanced',[n('support_top_z_distance','Top Z distance','mm',0,2),n('support_bottom_z_distance','Bottom Z distance','mm',0,2),n('support_base_pattern_spacing','Base pattern spacing','mm',0,10),n('support_interface_top_layers','Top interface layers','layers',0,20,1),n('support_object_xy_distance','Support/object XY distance','mm',0,10)])]:[])];
 case 'Others':return [group('Brim',[n('brim_width','Brim width','mm',0,30,.1)])];
 default:return [];
 }
}
