// Pinned Bambu Studio PartPlateList uses ceil(sqrt(count)) columns and a 20% gap.
// Meshes retain the original archive coordinates; only the disposable view moves.
export function nativePlateOrigin(index:number,count:number,width:number):[number,number,number]{
 const columns=Math.max(1,Math.ceil(Math.sqrt(count)));
 return [(index%columns)*width*1.2,-Math.floor(index/columns)*width*1.2||0,0];
}
export function displayPlateOrigin(index:number,count:number,width:number):[number,number,number]{
 const columns=Math.max(1,Math.ceil(Math.sqrt(count)));
 return [(index%columns)*(width+38),Math.floor(index/columns)*(width+38),0];
}
