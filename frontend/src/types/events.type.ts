export type OpenPanelEvent ={
  name:string,
  args : any
}

export type Events = {
    openPanel: OpenPanelEvent;
    closePanel: string;
  };
  