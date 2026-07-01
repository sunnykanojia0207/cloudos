import{b as t,d as a,e as n}from"./index-eBTYYIgn.js";/**
 * @license lucide-react v0.460.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const o=t("ExternalLink",[["path",{d:"M15 3h6v6",key:"1q9fwt"}],["path",{d:"M10 14 21 3",key:"gplh6r"}],["path",{d:"M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6",key:"a6xqqp"}]]);/**
 * @license lucide-react v0.460.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const c=t("GitCommitHorizontal",[["circle",{cx:"12",cy:"12",r:"3",key:"1v7zrd"}],["line",{x1:"3",x2:"9",y1:"12",y2:"12",key:"1dyftd"}],["line",{x1:"15",x2:"21",y1:"12",y2:"12",key:"oup4p8"}]]);function s(){return a({queryKey:["applications"],queryFn:async()=>{const e=await n.getResources("Application");return(e==null?void 0:e.items)??[]},refetchInterval:1e4})}function p(e){return a({queryKey:["applications",e],queryFn:async()=>await n.getResource("Application",e),enabled:!!e,refetchInterval:1e4})}export{o as E,c as G,p as a,s as u};
