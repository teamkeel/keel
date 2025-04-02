package flows

// // StartRun will start the scope's flow with the given input
// func StartRun(ctx context.Context, flow *proto.Flow, inputs any) (*Run, error) {
// 	if scope.Flow == nil {
// 		return nil, fmt.Errorf("invalid flow")
// 	}
// 	ctx, span := tracer.Start(ctx, "StartRun")
// 	defer span.End()

// 	span.SetAttributes(
// 		attribute.String("flow", scope.Flow.Name),
// 	)

// 	var jsonInputs JSONB
// 	if inputsMap, ok := inputs.(map[string]any); ok {
// 		jsonInputs = inputsMap
// 	}

// 	run := Run{
// 		Status: StatusNew,
// 		Input:  &jsonInputs,
// 		Name:   scope.Flow.Name,
// 	}

// 	database, err := db.GetDatabase(ctx)
// 	if err != nil {
// 		span.RecordError(err, trace.WithStackTrace(true))
// 		span.SetStatus(codes.Error, err.Error())
// 		return nil, err
// 	}

// 	result := database.GetDB().Create(&run)
// 	if result.Error != nil {
// 		span.RecordError(result.Error, trace.WithStackTrace(true))
// 		span.SetStatus(codes.Error, result.Error.Error())
// 		return nil, result.Error
// 	}

// 	span.SetAttributes(attribute.String("flowRunID", run.ID))

// 	// TODO: this will move to the orchestrator.  For now we are just running the flow synchronously.
// 	err = functions.CallFlow(
// 		ctx,
// 		scope.Flow,
// 		run.ID,
// 		inputs.(map[string]any),
// 	)
// 	if err != nil {
// 		span.RecordError(err, trace.WithStackTrace(true))
// 		span.SetStatus(codes.Error, err.Error())
// 		return nil, err
// 	}

// 	return &run, nil
// }
