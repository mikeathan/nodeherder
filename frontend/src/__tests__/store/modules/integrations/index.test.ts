import { createStore } from 'vuex';
import { IntegrationsModule, type ConnectionsState } from '@/store/modules/integrations';
import { connectionsApi, ConnectionError } from '@/services/integrations.service';
import type { ConnectionCatalog } from '@/types/integrations.type';
const catalog:ConnectionCatalog={available:true,availability_error:null,types:[],providers:[]};
afterEach(()=>jest.restoreAllMocks());
function fixture(){const logout=jest.fn();const store=createStore({modules:{integrations:IntegrationsModule,auth:{namespaced:true,state:()=>({}),mutations:{logout}}}});return {store,logout};}
test('a late response cannot repopulate private state after logout cleanup',async()=>{const {store}=fixture();let resolve!:(v:ConnectionCatalog)=>void;jest.spyOn(connectionsApi,'catalog').mockImplementation(()=>new Promise(done=>{resolve=done;}));jest.spyOn(connectionsApi,'instances').mockResolvedValue([]);const pending=store.dispatch('integrations/load');store.commit('integrations/clear');resolve(catalog);await pending;expect((store.state as {integrations:ConnectionsState}).integrations.catalog).toBeNull();expect((store.state as {integrations:ConnectionsState}).integrations.drafts).toEqual({});});
test('a late 401 from an earlier feature generation cannot sign out a new session',async()=>{const {store,logout}=fixture();let reject!:(e:unknown)=>void;jest.spyOn(connectionsApi,'catalog').mockImplementation(()=>new Promise((_resolve,fail)=>{reject=fail;}));jest.spyOn(connectionsApi,'instances').mockResolvedValue([]);const pending=store.dispatch('integrations/load');store.commit('integrations/clear');reject(new ConnectionError({code:'unauthorized',message:'Expired'},401));await expect(pending).rejects.toBeInstanceOf(ConnectionError);expect(logout).not.toHaveBeenCalled();});
