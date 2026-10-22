import React from 'react';
import {Fields} from './components/ui';
import {Recovery} from './generated/Recovery';
import {Product as Workflow} from './product';

class ScreenBoundary extends React.Component<{children?:React.ReactNode},{failed:boolean}> {
 state={failed:false};
 static getDerivedStateFromError(){return {failed:true};}
 render(){
  if(!this.state.failed)return this.props.children;
  return React.createElement(Fields.Provider,{value:{fields:{},setField:()=>{},disabled:false}},React.createElement(Recovery,{data:{},onAction:(action:string)=>{if(action==='prepare')location.hash='#';location.reload();}}));
 }
}
export function Product(props:any){return React.createElement(ScreenBoundary,{key:props.user.id},React.createElement(Workflow,props));}
