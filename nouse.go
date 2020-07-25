package obalance

//var mgo = omongo.NewMongoDB("mongodb://192.168.1.3:27017/?connect=direct", "balance")
//var mgo = omongo.NewMongoDB("mongodb://root:1314520Aa@dds-8vb9740883b210941.mongodb.zhangbei.rds.aliyuncs.com:3717,"+
//	"dds-8vb9740883b210942.mongodb.zhangbei.rds.aliyuncs.com:3717/admin?replicaSet=mgset-500744466", "balance")

//func (t *Trans) Trans() (interface{}, error) {
//	opts := options.Session().SetDefaultReadConcern(readconcern.Majority())
//	sess, err := mgo.MgoClient.StartSession(opts)
//	if err != nil {
//		return nil, err
//	}
//	defer sess.EndSession(context.Background())
//	txnOpts := options.Transaction().SetReadPreference(readpref.PrimaryPreferred())
//	result, err := sess.WithTransaction(context.Background(), func(sessCtx mongo.SessionContext) (interface{}, error) {
//		//插入logs，如果logs的id存在，就会插入失败
//		collLog := mgo.C("logs")
//		_, err := collLog.InsertOne(sessCtx, t)
//		if err != nil {
//			return nil, err
//		}
//		coll := mgo.C("balance")
//
//		//减少某个用户的余额和价格,当用户名为system时，代表从系统无限生成余额
//		if t.From != "system" {
//			fromTotal := join(t.Currency, "total")
//			fromType := join(t.Currency, t.From)
//			rstD := coll.FindOneAndUpdate(sessCtx,
//				bson.D{
//					{"bid", t.From},
//					{fromType, bson.M{"$gte": t.Amount - amountError}},
//				},
//				bson.M{"$inc": bson.M{
//					fromTotal: -t.Amount,
//					fromType:  -t.Amount,
//				}},
//				options.FindOneAndUpdate().SetUpsert(true).SetProjection(bson.M{t.Currency: 1}),
//			)
//			if rstD.Err() != nil {
//				if rstD.Err() == mongo.ErrNoDocuments {
//					return nil, errors.New("notEnoughBalance")
//				}
//				return nil, rstD.Err()
//			}
//			var fromMap map[string]float64
//			err = rstD.Decode(&fromMap)
//			if err != nil {
//				return nil, err
//			}
//			t.FromBalance = fromMap
//		}
//
//		//增加某个用户的余额
//		toAmount := t.Amount
//		if t.Fee > 0 {
//			getFee := t.Amount * t.Fee
//			//如果需要收手续费，那么增加对应的服务器账户
//			_, err = coll.Upsert(sessCtx, bson.M{"bid": t.Name}, bson.M{"$inc": bson.M{t.Currency: getFee}})
//			if err != nil {
//				return nil, err
//			}
//			toAmount = t.Amount - getFee
//		}
//		toTotal := join(t.Currency, "total")
//		toType := join(t.Currency, t.ToType)
//		rstI := coll.FindOneAndUpdate(sessCtx,
//			bson.M{"bid": t.To},
//			bson.M{"$inc": bson.M{
//				toTotal: toAmount,
//				toType:  toAmount,
//			}},
//			options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
//			options.FindOneAndUpdate().SetProjection(bson.M{t.Currency: 1, "_id": 0}),
//		)
//		if rstI.Err() != nil {
//			return nil, rstI.Err()
//		}
//		var toMap map[string]map[string]float64
//		err = rstI.Decode(&toMap)
//		if err != nil {
//			return nil, err
//		}
//		t.ToBalance = toMap[t.Currency]
//		t.Success = true
//		//减少余额
//		//增加余额
//		_, err = mgo.CDb("balance", "logs").UpdateOne(sessCtx, bson.M{"_id": t.ID}, bson.M{"$set": t})
//		if err != nil {
//			return nil, err
//		}
//		return t, nil
//	}, txnOpts)
//	return result, err
//}
//const amountError float64 = 0.000000001
