export type OpenPanelEvent ={
  name:string,
  args : any
}

need key value pair of events 

export type Events = {
    openPanel: OpenPanelEvent;
    closePanel: string;
  };
  