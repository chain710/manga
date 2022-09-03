"use strict";(self["webpackChunkview"]=self["webpackChunkview"]||[]).push([[801],{7801:function(t,e,i){i.r(e),i.d(e,{default:function(){return N}});var a=i(4562),r=i(9582),s=i(4886),n=i(2118),o=i(4324),l=i(5372),h=i(1444),c=i(4263),u=i(144),d=u.ZP.extend({name:"transitionable",props:{mode:String,origin:String,transition:String}}),p=i(5942),m=i(7678),b=(0,m.Z)(c.Z,h.Z,d).extend({name:"v-speed-dial",directives:{ClickOutside:p.Z},props:{direction:{type:String,default:"top",validator:t=>["top","right","bottom","left"].includes(t)},openOnHover:Boolean,transition:{type:String,default:"scale-transition"}},computed:{classes(){return{"v-speed-dial":!0,"v-speed-dial--top":this.top,"v-speed-dial--right":this.right,"v-speed-dial--bottom":this.bottom,"v-speed-dial--left":this.left,"v-speed-dial--absolute":this.absolute,"v-speed-dial--fixed":this.fixed,[`v-speed-dial--direction-${this.direction}`]:!0,"v-speed-dial--is-active":this.isActive}}},render(t){let e=[];const i={class:this.classes,directives:[{name:"click-outside",value:()=>this.isActive=!1}],on:{click:()=>this.isActive=!this.isActive}};if(this.openOnHover&&(i.on.mouseenter=()=>this.isActive=!0,i.on.mouseleave=()=>this.isActive=!1),this.isActive){let i=0;e=(this.$slots.default||[]).map(((e,a)=>!e.tag||"undefined"===typeof e.componentOptions||"v-btn"!==e.componentOptions.Ctor.options.name&&"v-tooltip"!==e.componentOptions.Ctor.options.name?(e.key=a,e):(i++,t("div",{style:{transitionDelay:.05*i+"s"},key:a},[e]))))}const a=t("transition-group",{class:"v-speed-dial__list",props:{name:this.transition,mode:this.mode,origin:this.origin,tag:"div"}},e);return t("div",i,[this.$slots.activator,a])}}),g=i(2082),f=function(){var t=this,e=t._self._c;return e("div",[e(n.Z,{attrs:{fluid:""}},[t.books.length>0?e(l.Z,{attrs:{length:t.pageCount,"total-visible":"8"},on:{input:t.jumpPage},model:{value:t.desiredPage,callback:function(e){t.desiredPage=e},expression:"desiredPage"}}):t._e(),t.books.length>0?e("item-browser",{attrs:{width:150,items:t.items,wrap:""}}):t._e(),t.isSetup&&0==t.books.length?e(r.Z,{staticClass:"mt-6",attrs:{elevation:"0"}},[e(s.EB,{staticClass:"d-flex justify-center align-center"},[e("h1",[e(o.Z,{attrs:{"x-large":"",color:"warning"}},[t._v("mdi-alert-decagram-outline")]),t._v(" "+t._s(t.$t("library.empty"))+" ")],1)])],1):t._e(),null!=t.library?e(b,{attrs:{bottom:"",right:"",absolute:"",fixed:"",direction:"top",transition:"slide-y-reverse-transition"},scopedSlots:t._u([{key:"activator",fn:function(){return[e(a.Z,{attrs:{color:"blue darken-2",dark:"",fab:""},model:{value:t.fab,callback:function(e){t.fab=e},expression:"fab"}},[t.fab?e(o.Z,[t._v("mdi-close")]):e(o.Z,[t._v("mdi-cog-outline")])],1)]},proxy:!0}],null,!1,94446779),model:{value:t.fab,callback:function(e){t.fab=e},expression:"fab"}},t._l(t.fabItems,(function(i,r){return e(a.Z,{key:r,attrs:{fab:"",dark:"",small:"",color:i.color},on:{click:i.onClick}},[e(g.Z,{attrs:{left:"","nudge-left":"5","open-delay":"500"},scopedSlots:t._u([{key:"activator",fn:function({on:a,attrs:r}){return[e(o.Z,t._g(t._b({},"v-icon",r,!1),a),[t._v(t._s(i.icon))])]}}],null,!0)},[e("span",[t._v(t._s(i.tip))])])],1)})),1):t._e(),e("confirm-dialog",{attrs:{title:t.confirm.title,body:t.confirm.body,type:"error","confirm-func":t.confirm.do},model:{value:t.confirm.enabled,callback:function(e){t.$set(t.confirm,"enabled",e)},expression:"confirm.enabled"}}),e("library-edit-dialog",{attrs:{library:t.library},on:{updated:t.$hub.syncLibraries},model:{value:t.showLibraryEdit,callback:function(e){t.showLibraryEdit=e},expression:"showLibraryEdit"}})],1)],1)},v=[],y=i(6486),$=i.n(y),k=i(3545),_=(i(7393),i(2240)),x=i(573),I=i(596),C=I.Z.extend({name:"v-checkbox",props:{indeterminate:Boolean,indeterminateIcon:{type:String,default:"$checkboxIndeterminate"},offIcon:{type:String,default:"$checkboxOff"},onIcon:{type:String,default:"$checkboxOn"}},data(){return{inputIndeterminate:this.indeterminate}},computed:{classes(){return{...x.Z.options.computed.classes.call(this),"v-input--selection-controls":!0,"v-input--checkbox":!0,"v-input--indeterminate":this.inputIndeterminate}},computedIcon(){return this.inputIndeterminate?this.indeterminateIcon:this.isActive?this.onIcon:this.offIcon},validationState(){if(!this.isDisabled||this.inputIndeterminate)return this.hasError&&this.shouldValidate?"error":this.hasSuccess?"success":null!==this.hasColor?this.computedColor:void 0}},watch:{indeterminate(t){this.$nextTick((()=>this.inputIndeterminate=t))},inputIndeterminate(t){this.$emit("update:indeterminate",t)},isActive(){this.indeterminate&&(this.inputIndeterminate=!1)}},methods:{genCheckbox(){const{title:t,...e}=this.attrs$;return this.$createElement("div",{staticClass:"v-input--selection-controls__input"},[this.$createElement(_.Z,this.setTextColor(this.validationState,{props:{dense:this.dense,dark:this.dark,light:this.light}}),this.computedIcon),this.genInput("checkbox",{...e,"aria-checked":this.inputIndeterminate?"mixed":this.isActive.toString()}),this.genRipple(this.setTextColor(this.rippleState))])},genDefaultSlot(){return[this.genCheckbox(),this.genLabel()]}}}),S=i(266),Z=i(4061),L=i(1713),w=i(3687),A=function(){var t=this,e=t._self._c;return e(Z.Z,{attrs:{"max-width":"450"},model:{value:t.modal,callback:function(e){t.modal=e},expression:"modal"}},[e(r.Z,[e(s.EB,{class:t.titleClass},[t._v(t._s(t.title))]),e(s.ZB,[e(n.Z,{attrs:{fluid:""}},[e(L.Z,[t.body?e(S.Z,[t._t("default",(function(){return[t._v(t._s(t.body))]}))],2):t._e()],1),t.confirmText?e(L.Z,[e(S.Z,[e(C,{attrs:{color:t.color},scopedSlots:t._u([{key:"label",fn:function(){return[t._v(" "+t._s(t.confirmText)+" ")]},proxy:!0}],null,!1,1982615865),model:{value:t.confirmation,callback:function(e){t.confirmation=e},expression:"confirmation"}})],1)],1):t._e()],1)],1),e(s.h7,[e(w.Z),e(a.Z,{attrs:{text:""},on:{click:t.dialogCancel}},[t._v(t._s(t.buttonCancel||t.$t("global.cancel")))]),e(a.Z,{attrs:{color:t.color,loading:t.loading,disabled:t.confirmText&&!t.confirmation},on:{click:t.dialogConfirm}},[t._v(" "+t._s(t.buttonConfirm||t.$t("global.confirm"))+" ")])],1)],1)],1)},V=[],D={data:function(){return{modal:this.value,confirmation:!1,loading:!1}},props:{value:Boolean,title:{type:String,required:!0},body:{type:String,required:!1},confirmText:{type:String,required:!1},confirmFunc:{type:Function,required:!1},buttonCancel:{type:String,required:!1},buttonConfirm:{type:String,required:!1},type:{type:String,default:"primary"}},watch:{value(t){this.modal=t},modal(t){!t&&this.dialogCancel()}},methods:{dialogCancel(){this.confirmation=!1,this.$emit("input",!1)},async dialogConfirm(){this.confirmFunc||(this.$emit("confirm"),this.$emit("input",!1));try{this.loading=!0,await this.confirmFunc()}catch(t){console.log("confirm func error",t)}finally{this.loading=!1,this.$emit("confirm"),this.$emit("input",!1)}}},computed:{color(){return this.type},titleClass(){switch(this.type){case"error":return"red white--text";default:return this.type}}}},B=D,E=i(1001),P=(0,E.Z)(B,A,V,!1,null,"577e4cf6",null),M=P.exports,q=i(8462),z={components:{ItemBrowser:k.Z,ConfirmDialog:M,LibraryEditDialog:q.Z},data:function(){return{dataPage:1,desiredPage:1,totalElements:0,books:[],pageSize:20,fab:!1,library:null,fabItems:[{icon:"mdi-delete",onClick:this.confirmDeleteLibrary,tip:this.$t("library.fab.delete"),color:"red"},{icon:"mdi-pencil",onClick:this.editLibrary,tip:this.$t("library.fab.edit"),color:"green"},{icon:"mdi-magnify-scan",onClick:this.scanLibrary,tip:this.$t("library.fab.scan"),color:"indigo"}],confirm:{enabled:!1,title:"",body:""},showLibraryEdit:!1,isSetup:!1}},props:{libraryID:{}},async mounted(){const t=166,e=Math.floor(window.innerWidth/t);this.pageSize=Math.max(e,Math.floor(50/e)*e),await this.setup(this.libraryID,this.$route.query.page),this.isSetup=!0},methods:{jumpPage(t){if(t==this.dataPage)return;let e=$().clone(this.$route.query);e["page"]=t,this.$router.replace({name:"libraries",params:this.$route.params,query:e})},async syncLibBooks(t){let e={};$().parseInt(t)&&(e["lib"]=t),e["filter"]="with_progress_relax",e["offset"]=this.pageSize*(this.desiredPage-1),e["limit"]=this.pageSize;try{let t=await this.$service.listBooks(e);this.books=t.data.data.books,this.totalElements=t.data.data.count,this.dataPage=this.desiredPage}catch(i){this.$nerror("list_book",i),console.error(`list book error: ${i}`)}},async syncLibrary(t){if($().parseInt(t))try{let e=await this.$service.getLibrary(t);this.library=e.data.data}catch(e){this.$nerror("list_library",e),console.error(`get library error: ${e}`)}else this.library=null},editLibrary(){this.showLibraryEdit=!0},async scanLibrary(){try{let t=await this.$service.scanLibrary(this.libraryID);console.log("scan library resp",t),this.$ninfo("scan_library")}catch(t){console.log("scan library error: ",t),this.$nerror("scan_library",t)}},confirmDeleteLibrary(){this.confirm.title=this.$t("dialog.delete_library.title",{name:this.library.name}),this.confirm.body=this.$t("dialog.delete_library.body",{path:this.library.path}),this.confirm.do=this.deleteLibrary,this.confirm.enabled=!0},async deleteLibrary(){try{let t=await this.$service.deleteLibrary(this.library.id);console.log(`delete library ${this.library.id}`,t),this.$ninfo("delete_library")}catch(t){this.$nerror("delete_library",t),console.error(`delete library error ${this.library.id}`,t)}try{await this.$hub.syncLibraries(),this.$router.replace({name:"libraries",params:{libraryID:this.$LIBRARY_ID_ALL}}).catch((t=>{console.debug("replace lib error",t)}))}catch(t){console.error("navigate to libraries error",t)}},async setup(t,e){const i=$().parseInt(e);this.desiredPage=i||1;const a=Promise.all([this.syncLibBooks(t),this.syncLibrary(t)]);await this.$hub.addTask(a),this.$emit("main-enter",{name:"library",value:{id:this.libraryID,name:this.library?this.library.name:this.$t("library.all"),count:this.totalElements}})}},computed:{items(){let t=[];for(let e of this.books)t.push(this.$convertBook(e));return t},pageCount(){return Math.ceil(this.totalElements/this.pageSize)}},async beforeRouteUpdate(t,e,i){t.params.libraryID==e.params.libraryID&&t.query.page==e.query.page||await this.setup(t.params.libraryID,t.query.page),i()}},R=z,T=(0,E.Z)(R,f,v,!1,null,"24615a5e",null),N=T.exports},7393:function(){},5372:function(t,e,i){i.d(e,{Z:function(){return h}});var a=i(2240),r=i(6746),s=i(6878),n=i(7756),o=i(6669),l=i(7678),h=(0,l.Z)(s.Z,(0,n.Z)({onVisible:["init"]}),o.Z).extend({name:"v-pagination",directives:{Resize:r.Z},props:{circle:Boolean,disabled:Boolean,length:{type:Number,default:0,validator:t=>t%1===0},nextIcon:{type:String,default:"$next"},prevIcon:{type:String,default:"$prev"},totalVisible:[Number,String],value:{type:Number,default:0},pageAriaLabel:{type:String,default:"$vuetify.pagination.ariaLabel.page"},currentPageAriaLabel:{type:String,default:"$vuetify.pagination.ariaLabel.currentPage"},previousAriaLabel:{type:String,default:"$vuetify.pagination.ariaLabel.previous"},nextAriaLabel:{type:String,default:"$vuetify.pagination.ariaLabel.next"},wrapperAriaLabel:{type:String,default:"$vuetify.pagination.ariaLabel.wrapper"}},data(){return{maxButtons:0,selected:null}},computed:{classes(){return{"v-pagination":!0,"v-pagination--circle":this.circle,"v-pagination--disabled":this.disabled,...this.themeClasses}},items(){const t=parseInt(this.totalVisible,10);if(0===t||isNaN(this.length)||this.length>Number.MAX_SAFE_INTEGER)return[];const e=Math.min(Math.max(0,t)||this.length,Math.max(0,this.maxButtons)||this.length,this.length);if(this.length<=e)return this.range(1,this.length);const i=e%2===0?1:0,a=Math.floor(e/2),r=this.length-a+1+i;if(this.value>a&&this.value<r){const t=1,e=this.length,r=this.value-a+2,s=this.value+a-2-i,n=r-1===t+1?2:"...",o=s+1===e-1?s+1:"...";return[1,n,...this.range(r,s),o,this.length]}if(this.value===a){const t=this.value+a-1-i;return[...this.range(1,t),"...",this.length]}if(this.value===r){const t=this.value-a+1;return[1,"...",...this.range(t,this.length)]}return[...this.range(1,a),"...",...this.range(r,this.length)]}},watch:{value(){this.init()}},beforeMount(){this.init()},methods:{init(){this.selected=null,this.onResize(),this.$nextTick(this.onResize),setTimeout((()=>this.selected=this.value),100)},onResize(){const t=this.$el&&this.$el.parentElement?this.$el.parentElement.clientWidth:window.innerWidth;this.maxButtons=Math.floor((t-96)/42)},next(t){t.preventDefault(),this.$emit("input",this.value+1),this.$emit("next")},previous(t){t.preventDefault(),this.$emit("input",this.value-1),this.$emit("previous")},range(t,e){const i=[];t=t>0?t:1;for(let a=t;a<=e;a++)i.push(a);return i},genIcon(t,e,i,r,s){return t("li",[t("button",{staticClass:"v-pagination__navigation",class:{"v-pagination__navigation--disabled":i},attrs:{disabled:i,type:"button","aria-label":s},on:i?{}:{click:r}},[t(a.Z,[e])])])},genItem(t,e){const i=e===this.value&&(this.color||"primary"),a=e===this.value,r=a?this.currentPageAriaLabel:this.pageAriaLabel;return t("button",this.setBackgroundColor(i,{staticClass:"v-pagination__item",class:{"v-pagination__item--active":e===this.value},attrs:{type:"button","aria-current":a,"aria-label":this.$vuetify.lang.t(r,e)},on:{click:()=>this.$emit("input",e)}}),[e.toString()])},genItems(t){return this.items.map(((e,i)=>t("li",{key:i},[isNaN(Number(e))?t("span",{class:"v-pagination__more"},[e.toString()]):this.genItem(t,e)])))},genList(t,e){return t("ul",{directives:[{modifiers:{quiet:!0},name:"resize",value:this.onResize}],class:this.classes},e)}},render(t){const e=[this.genIcon(t,this.$vuetify.rtl?this.nextIcon:this.prevIcon,this.value<=1,this.previous,this.$vuetify.lang.t(this.previousAriaLabel)),this.genItems(t),this.genIcon(t,this.$vuetify.rtl?this.prevIcon:this.nextIcon,this.value>=this.length,this.next,this.$vuetify.lang.t(this.nextAriaLabel))];return t("nav",{attrs:{role:"navigation","aria-label":this.$vuetify.lang.t(this.wrapperAriaLabel)}},[this.genList(t,e)])}})},596:function(t,e,i){i.d(e,{Z:function(){return c}});var a=i(573),r=i(7069),s=i(144),n=s.ZP.extend({name:"rippleable",directives:{ripple:r.Z},props:{ripple:{type:[Boolean,Object],default:!0}},methods:{genRipple(t={}){return this.ripple?(t.staticClass="v-input--selection-controls__ripple",t.directives=t.directives||[],t.directives.push({name:"ripple",value:{center:!0}}),this.$createElement("div",t)):null}}}),o=i(6174),l=i(7678);function h(t){t.preventDefault()}var c=(0,l.Z)(a.Z,n,o.Z).extend({name:"selectable",model:{prop:"inputValue",event:"change"},props:{id:String,inputValue:null,falseValue:null,trueValue:null,multiple:{type:Boolean,default:null},label:String},data(){return{hasColor:this.inputValue,lazyValue:this.inputValue}},computed:{computedColor(){if(this.isActive)return this.color?this.color:this.isDark&&!this.appIsDark?"white":"primary"},isMultiple(){return!0===this.multiple||null===this.multiple&&Array.isArray(this.internalValue)},isActive(){const t=this.value,e=this.internalValue;return this.isMultiple?!!Array.isArray(e)&&e.some((e=>this.valueComparator(e,t))):void 0===this.trueValue||void 0===this.falseValue?t?this.valueComparator(t,e):Boolean(e):this.valueComparator(e,this.trueValue)},isDirty(){return this.isActive},rippleState(){return this.isDisabled||this.validationState?this.validationState:void 0}},watch:{inputValue(t){this.lazyValue=t,this.hasColor=t}},methods:{genLabel(){const t=a.Z.options.methods.genLabel.call(this);return t?(t.data.on={click:h},t):t},genInput(t,e){return this.$createElement("input",{attrs:Object.assign({"aria-checked":this.isActive.toString(),disabled:this.isDisabled,id:this.computedId,role:t,type:t},e),domProps:{value:this.value,checked:this.isActive},on:{blur:this.onBlur,change:this.onChange,focus:this.onFocus,keydown:this.onKeydown,click:h},ref:"input"})},onClick(t){this.onChange(),this.$emit("click",t)},onChange(){if(!this.isInteractive)return;const t=this.value;let e=this.internalValue;if(this.isMultiple){Array.isArray(e)||(e=[]);const i=e.length;e=e.filter((e=>!this.valueComparator(e,t))),e.length===i&&e.push(t)}else e=void 0!==this.trueValue&&void 0!==this.falseValue?this.valueComparator(e,this.trueValue)?this.falseValue:this.trueValue:t?this.valueComparator(e,t)?null:t:!e;this.validate(!0,e),this.internalValue=e,this.hasColor=e},onFocus(t){this.isFocused=!0,this.$emit("focus",t)},onBlur(t){this.isFocused=!1,this.$emit("blur",t)},onKeydown(t){}}})}}]);
//# sourceMappingURL=801.592548ba.js.map