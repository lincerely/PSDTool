!function(r,t){"object"==typeof exports&&"object"==typeof module?module.exports=t():"function"==typeof define&&define.amd?define([],t):"object"==typeof exports?exports.LZ4=t():r.LZ4=t()}(this,function(){return function(r){function t(o){if(f[o])return f[o].exports;var i=f[o]={exports:{},id:o,loaded:!1};return r[o].call(i.exports,i,i.exports,t),i.loaded=!0,i.exports}var f={};return t.m=r,t.c=f,t.p="",t(0)}([function(r,t){
// lz4-ts @license BSD-3-Clause / Copyright (c) 2015, Pierre Curto / 2016, oov. All rights reserved.
"use strict";function f(r,t){var f=r>>>16&65535,o=65535&r,i=t>>>16&65535,e=65535&t;return o*e+(f*e+o*i<<16>>>0)|0}function o(r,t){return r[t+3]|r[t+2]<<8|r[t+1]<<16|r[t]<<24}function i(r,t,f,o,i){for(var e=0;e<i;++e)r[f++]=t[o++]}function e(r){var t=new Uint8Array(r),f=t.length;if(0===f)return 0;for(var o=0,i=0;;){var e=t[o]>>4,n=15&t[o];if(++o===f)throw u;if(e>0){if(15===e){for(;255===t[o];)if(e+=255,++o===f)throw u;if(e+=t[o],++o===f)throw u}if(i+=e,o+=e,o>=f)return i}if(o+=2,o>=f)throw u;var a=t[o-2]|t[o-1]<<8;if(i-a<0||0===a)throw u;if(15===n){for(;255===t[o];)if(n+=255,++o===f)throw u;if(n+=t[o],++o===f)throw u}for(n+=4;n>=a;n-=a)i+=a;i+=n}}function n(r,t){var f=new Uint8Array(r),o=new Uint8Array(t),e=f.length,n=o.length;if(0===e)return 0;for(var a=0,h=0;;){var w=f[a]>>4,c=15&f[a];if(++a===e)throw u;if(w>0){if(15===w){for(;255===f[a];)if(w+=255,++a===e)throw u;if(w+=f[a],++a===e)throw u}if(n-h<w||a+w>e)throw v;if(i(o,f,h,a,w),h+=w,a+=w,a>=e)return h}if(a+=2,a>=e)throw u;var l=f[a-2]|f[a-1]<<8;if(h-l<0||0===l)throw u;if(15===c){for(;255===f[a];)if(c+=255,++a===e)throw u;if(c+=f[a],++a===e)throw u}if(c+=4,n-h<=c)throw v;for(;c>=l;c-=l)i(o,o,h,h-l,l),h+=l;i(o,o,h,h-l,c),h+=c}}function a(r){return r+(r/255|0)+16}function h(r,t,f){var e=new Uint8Array(r),n=new Uint8Array(t),a=e.length-m,h=n.length;if(a<=0||0===h||f>=a)return 0;for(var w=0,u=0,s=new Uint32Array(y);w<f;){var p=x(o(e,w),A)>>>g;s[p]=++w}for(var d=w,b=1<<U;w<a-c;){var p=x(o(e,w),A)>>>g,k=s[p]-1;if(s[p]=w+1,k<0||w-k>>l>0||e[k]!==e[w]||e[k+1]!==e[w+1]||e[k+2]!==e[w+2]||e[k+3]!==e[w+3])w+=b>>U,++b;else{b=1<<U;var B=w-d,j=w-k;w+=c;for(var L=w;w<=a&&e[w]===e[w-j];)w++;if(L=w-L,L<15?n[u]=L:n[u]=15,B<15)n[u]|=B<<4;else{if(n[u]|=240,++u===h)throw v;for(var E=B-15;E>=255;E-=255)if(n[u]=255,++u===h)throw v;n[u]=255&E}if(++u===h)throw v;if(u+B>=h)throw v;if(i(n,e,u,d,B),u+=B,d=w,u+=2,u>=h)throw v;if(n[u-2]=j,n[u-1]=j>>8,L>=15){for(L-=15;L>=255;L-=255)if(n[u]=255,++u===h)throw v;if(n[u]=L,++u===h)throw v}}}if(0===d)return 0;var M=e.length-d;if(M<15)n[u]=M<<4;else{if(n[u]=240,++u===h)throw v;for(M-=15;M>=255;M-=255)if(n[u]=255,++u===h)throw v;n[u]=M}if(++u===h)throw v;var Z=e.length-d,C=u+Z;if(C>h)throw v;return C>=a?0:(i(n,e,u,d,Z),u+=Z)}function w(r,t,f){var e=new Uint8Array(r),n=new Uint8Array(t),a=e.length-m,h=n.length;if(a<=0||0===h||f>=a)return 0;for(var w=0,u=0,l=new Uint32Array(y),d=new Uint32Array(s);w<f;){var U=x(o(e,w),A)>>>g;d[w&p]=l[U],l[U]=++w}for(var b=w;w<a-c;){for(var U=x(o(e,w),A)>>>g,k=0,B=0,j=l[U]-1;j>0&&j>w-s;j=d[j&p]-1)if(e[j+k]===e[w+k])for(var L=0;;++L)if(e[j+L]!==e[w+L]||w+L>a){k<L&&L>=c&&(k=L,B=w-j);break}if(d[w&p]=l[U],l[U]=w+1,0!==k){for(var E=w+1,L=w+k;E<L;){var M=x(o(e,E),A)>>>g;d[E&p]=l[M],l[M]=++E}var Z=w-b;if(w+=k,k-=c,k<15?n[u]=k:n[u]=15,Z<15)n[u]|=Z<<4;else{if(n[u]|=240,++u===h)throw v;for(var C=Z-15;C>=255;C-=255)if(n[u]=255,++u===h)throw v;n[u]=255&C}if(++u===h)throw v;if(u+Z>=h)throw v;if(i(n,e,u,b,Z),u+=Z,b=w,u+=2,u>=h)throw v;if(n[u-2]=B,n[u-1]=B>>8,k>=15){for(k-=15;k>=255;k-=255)if(n[u]=255,++u===h)throw v;if(n[u]=k,++u===h)throw v}}else++w}if(0===b)return 0;var H=e.length-b;if(H<15)n[u]=H<<4;else{if(n[u]=240,++u===h)throw v;for(H-=15;H>=255;H-=255)if(n[u]=255,++u===h)throw v;n[u]=H}if(++u===h)throw v;var q=e.length-b,z=u+q;if(z>h)throw v;return z>=a?0:(i(n,e,u,b,q),u+=q)}var u=new Error("invalid source"),v=new Error("short buffer"),c=4,l=16,s=1<<l,p=s-1,d=16,y=1<<d,g=8*c-d,m=8+c,U=6,A=-1640531535,x=Math.imul?Math.imul:f;t.calcUncompressedLen=e,t.uncompressBlock=n,t.compressBlockBound=a,t.compressBlock=h,t.compressBlockHC=w}])});