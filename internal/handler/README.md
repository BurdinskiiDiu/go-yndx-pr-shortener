/*func (hn *Handlers) DelURLSBatch(inpChnl []postgresql.URLsForDel) {
	ctx := context.TODO()
	err := hn.US.DeleteUserURLS(ctx, inpChnl)
	if err != nil {
		hn.logger.Debug("error while del urls:" + err.Error())
	}
	//ticker := time.NewTicker(5 * time.Second)
	//delURLsSlc := make([]postgresql.URLsForDel, 0)
	/*for {
	select {
	case delURL := <-inpChnl:
		err := hn.US.DeleteUserURLS(ctx, delURL)
		if err != nil {
			hn.logger.Debug("error while del urls:" + err.Error())
		}
	default:
		continue
	}*/
	/*select {
	case delURL := <-inpChnl:
		delURLsSlc = append(delURLsSlc, delURL)
	case <-ticker.C:
		if len(delURLsSlc) == 0 {
			continue
		}
		err := hn.US.DeleteUserURLS(ctx, delURLsSlc)
		if err != nil {
			hn.logger.Debug("error while del urls:" + err.Error())
			continue
		}
		delURLsSlc = nil
	}*/
}

//}