package worker

type Work func() WorkResult

type WorkResult interface{}
