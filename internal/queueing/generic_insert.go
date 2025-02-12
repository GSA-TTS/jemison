package queueing

import (
	"context"
	"strings"

	"github.com/GSA-TTS/jemison/internal/common"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"go.uber.org/zap"
)

type QSHP struct {
	Queue      string
	Scheme     string
	Host       string
	Path       string
	IsFull     bool
	IsHallPass bool
	Filename   string
	RawData    string
}

//nolint:revive
func commonCommit(qshp QSHP, ctx context.Context, tx pgx.Tx) {
	if err := tx.Commit(ctx); err != nil {
		err = tx.Rollback(ctx)
		if err != nil {
			zap.L().Error("cannot roll back commit")
		}

		zap.L().Fatal("cannot commit insert tx",
			zap.String("host", qshp.Host),
			zap.String("path", qshp.Path),
			zap.String("err", err.Error()))
	}
}

//nolint:cyclop,funlen,gocognit
func Enqueue(chQSHP <-chan QSHP) {
	// Can we leave one connection open for the entire life of a
	// service? Maybe. Maybe not.
	_, pool, _ := common.CommonQueueInit()
	defer pool.Close()

	client, err := river.NewClient(riverpgxv5.New(pool), &river.Config{})
	if err != nil {
		zap.L().Error("could not create river client",
			zap.String("error", err.Error()))
	}

	for {
		qshp := <-chQSHP
		ctx, tx := common.CtxTx(pool)

		var queueToMatch string
		if strings.HasPrefix(qshp.Queue, "fetch") {
			queueToMatch = "fetch"
		} else {
			queueToMatch = qshp.Queue
		}

		switch queueToMatch {
		case "entree":
			_, err := client.InsertTx(ctx, tx, common.EntreeArgs{
				Scheme:    qshp.Scheme,
				Host:      qshp.Host,
				Path:      qshp.Path,
				FullCrawl: qshp.IsFull,
				HallPass:  qshp.IsHallPass,
			}, &river.InsertOpts{Queue: qshp.Queue})
			if err != nil {
				zap.L().Error("cannot insert into queue entree")
			}

			commonCommit(qshp, ctx, tx)

		case "extract":
			_, err := client.InsertTx(ctx, tx, common.ExtractArgs{
				Scheme: qshp.Scheme,
				Host:   qshp.Host,
				Path:   qshp.Path,
			}, &river.InsertOpts{Queue: qshp.Queue})
			if err != nil {
				zap.L().Error("cannot insert into queue extract")
			}

			commonCommit(qshp, ctx, tx)

		case "fetch":
			_, err = client.InsertTx(ctx, tx, common.FetchArgs{
				Scheme: qshp.Scheme,
				Host:   qshp.Host,
				Path:   qshp.Path,
			}, &river.InsertOpts{Queue: qshp.Queue})
			if err != nil {
				zap.L().Error("cannot insert into queue fetch")
			}

			commonCommit(qshp, ctx, tx)

		case "collect":
			zap.L().Debug("handling collect queue insertion",
				zap.String("scheme", qshp.Scheme),
				zap.String("host", qshp.Host),
				zap.String("path", qshp.Path),
				zap.String("rawData", qshp.RawData),
				zap.Bool("full", qshp.IsFull),
				zap.Bool("hallpass", qshp.IsHallPass))

			_, err := client.InsertTx(ctx, tx, common.CollectArgs{
				Scheme:    qshp.Scheme,
				Host:      qshp.Host,
				Path:      qshp.Path,
				JSON:      qshp.RawData,
				FullCrawl: qshp.IsFull,
				HallPass:  qshp.IsHallPass,
			}, &river.InsertOpts{Queue: qshp.Queue})
			if err != nil {
				zap.L().Error("cannot insert into queue collect",
					zap.String("host", qshp.Host), zap.String("path", qshp.Path),
					zap.String("error", err.Error()))
			}

			commonCommit(qshp, ctx, tx)

		case "pack":
			_, err = client.InsertTx(ctx, tx, common.PackArgs{
				Scheme: qshp.Scheme,
				Host:   qshp.Host,
				Path:   qshp.Path,
			}, &river.InsertOpts{Queue: qshp.Queue})
			if err != nil {
				zap.L().Error("cannot insert into queue pack")
			}

			commonCommit(qshp, ctx, tx)

		case "serve":
			_, err := client.InsertTx(ctx, tx, common.ServeArgs{
				Filename: qshp.Filename,
			}, &river.InsertOpts{Queue: qshp.Queue})
			if err != nil {
				zap.L().Error("cannot insert into queue serve")
			}

			commonCommit(qshp, ctx, tx)

		case "walk":
			if qshp.Queue != "walk" {
				zap.L().Error("found non-walk job coming to the walk queue",
					zap.String("host", qshp.Host), zap.String("path", qshp.Path))
			}

			_, err := client.InsertTx(ctx, tx, common.WalkArgs{
				Scheme: qshp.Scheme,
				Host:   qshp.Host,
				Path:   qshp.Path,
			}, &river.InsertOpts{Queue: "walk"})
			if err != nil {
				zap.L().Error("cannot insert into queue walk")
			}

			commonCommit(qshp, ctx, tx)

		default:
			zap.L().Error("unknown common enqueue", zap.String("queue", qshp.Queue))
		}
	}
}
