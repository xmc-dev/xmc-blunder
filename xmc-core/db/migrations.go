package db

import (
	"time"

	"github.com/go-gormigrate/gormigrate"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"github.com/xmc-dev/xmc/xmc-core/db/models/attachment"
	"github.com/xmc-dev/xmc/xmc-core/db/models/page"
	"github.com/xmc-dev/xmc/xmc-core/db/models/problem"
	"github.com/xmc-dev/xmc/xmc-core/db/models/submission"
	"github.com/xmc-dev/xmc/xmc-core/db/models/tasklist"
)

func (d *Datastore) Migrate() error {
	opts := gormigrate.DefaultOptions
	m := gormigrate.New(d.db, opts, []*gormigrate.Migration{
		{
			ID: "201712240001",
			Migrate: func(tx *gorm.DB) error {
				type Attachment struct {
					ID          uuid.UUID `gorm:"primary_key;type:uuid;default:uuid_generate_v1mc()"`
					S3Object    string
					Description string
					ObjectID    string `gorm:"unique_index:idx_object_id_filename"`
					Filename    string `gorm:"unique_index:idx_object_id_filename"`
				}
				return tx.AutoMigrate(&Attachment{}).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.DropTable("attachments").Error
			},
		},
		{
			ID: "201712240015",
			Migrate: func(tx *gorm.DB) error {
				type Grader struct {
					ID           uuid.UUID `gorm:"primary_key;type:uuid;default:uuid_generate_v1mc()"`
					AttachmentID uuid.UUID `gorm:"type:uuid"`
					Language     string
					Name         string `gorm:"unique"`
				}
				type Dataset struct {
					ID          uuid.UUID `gorm:"primary_key;type:uuid;default:uuid_generate_v1mc()"`
					GraderID    uuid.UUID `gorm:"type:uuid"`
					Description string
					TimeLimit   time.Duration
					MemoryLimit int32
				}
				type TestCase struct {
					ID                 uuid.UUID `gorm:"primary_key;type:uuid;default:uuid_generate_v1mc()"`
					DatasetID          uuid.UUID `gorm:"type:uuid"`
					Number             int32
					InputAttachmentID  uuid.UUID `gorm:"type:uuid"`
					OutputAttachmentID uuid.UUID `gorm:"type:uuid"`
				}
				type Task struct {
					ID        uuid.UUID `gorm:"primary_key;type:uuid;default:uuid_generate_v1mc()"`
					DatasetID uuid.UUID `gorm:"type:uuid"`
					Name      string
					// Internal description for administrators
					Description string
				}
				err := tx.AutoMigrate(&Grader{}, &Dataset{}, &TestCase{}, &Task{}).Error
				if err != nil {
					return err
				}

				err = tx.Model(&Dataset{}).AddForeignKey("grader_id", "graders(id)", "RESTRICT", "CASCADE").Error
				if err != nil {
					return err
				}

				return tx.Model(&Task{}).AddForeignKey("dataset_id", "datasets(id)", "RESTRICT", "CASCADE").Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.DropTable("graders", "datasets", "test_cases", "tasks").Error
			},
		},
		{
			// add unique index to TestCase
			ID: "201712240030",
			Migrate: func(tx *gorm.DB) error {
				return tx.Model(&problem.TestCase{}).AddUniqueIndex("idx_dataset_id_number", "dataset_id", "number").Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Model(&problem.TestCase{}).RemoveIndex("idx_dataset_id_number").Error
			},
		},
		{
			ID: "201712240045",
			Migrate: func(tx *gorm.DB) error {
				return tx.Model(&problem.TestCase{}).AddForeignKey("dataset_id", "datasets(id)", "CASCADE", "CASCADE").Error
			},
			Rollback: func(tx *gorm.DB) error {
				return rfk(tx, "test_cases", "dataset_id", "datasets(id)")
			},
		},
		{
			// add dataset's name field
			ID: "201712300015",
			Migrate: func(tx *gorm.DB) error {
				type Dataset struct {
					Name string `gorm:"unique_index;"`
				}
				return tx.AutoMigrate(&Dataset{}).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Model(&problem.Dataset{}).DropColumn("name").Error
			},
		},
		{
			ID: "201712300045",
			Migrate: func(tx *gorm.DB) error {
				return tx.Model(&problem.Task{}).AddUniqueIndex("idx_tasks_name", "name").Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Model(&problem.Task{}).RemoveIndex("idx_tasks_name").Error
			},
		},
		{
			ID: "201712300060",
			Migrate: func(tx *gorm.DB) error {
				return tx.Model(&problem.Grader{}).AddUniqueIndex("idx_graders_name", "name").Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Model(&problem.Task{}).RemoveIndex("idx_graders_name").Error
			},
		},
		{
			ID: "201712300075",
			Migrate: func(tx *gorm.DB) error {
				return tx.Exec("ALTER INDEX idx_object_id_filename RENAME TO idx_attachments_object_id_filename").
					Exec("ALTER INDEX idx_dataset_id_number RENAME TO idx_test_cases_dataset_id_number").Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Exec("ALTER INDEX idx_attachments_object_id_filename RENAME TO idx_object_id_filename").
					Exec("ALTER INDEX idx_test_cases_dataset_id_number RENAME TO idx_dataset_id_number").Error
			},
		},
		{
			ID: "201701030015",
			Migrate: func(tx *gorm.DB) error {
				type Task struct {
					InputFile  string
					OutputFile string
				}
				return tx.AutoMigrate(&Task{}).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Table("tasks").DropColumn("input_file").DropColumn("output_file").Error
			},
		},
		{
			ID: "201701040015",
			Migrate: func(tx *gorm.DB) error {
				type Submission struct {
					ID           uuid.UUID `gorm:"primary_key;type:uuid;default:uuid_generate_v1mc()"`
					TaskID       uuid.UUID `gorm:"type:uuid;index"`
					DatasetID    uuid.UUID `gorm:"type:uuid"`
					AttachmentID uuid.UUID `gorm:"type:uuid"`
					EvalID       string
					Language     string
					CreatedAt    time.Time
					FinishedAt   *time.Time
					State        int32
				}
				type SubmissionResult struct {
					SubmissionID       uuid.UUID `gorm:"type:uuid;primary_key"`
					ErrorMessage       string
					CompilationMessage string
				}
				type TestResult struct {
					SubmissionID  uuid.UUID `gorm:"type:uuid;primary_key"`
					TestNo        int32     `gorm:"primary_key"`
					Score         int32
					GraderMessage string
					Memory        int32
					Time          time.Duration
				}

				err := tx.AutoMigrate(&TestResult{}, &SubmissionResult{}, &Submission{}).Error
				if err != nil {
					return err
				}
				err = tx.Table("submission_results").
					AddForeignKey("submission_id", "submissions(id)", "CASCADE", "CASCADE").Error
				if err != nil {
					return err
				}
				err = tx.Table("test_results").
					AddForeignKey("submission_id", "submissions(id)", "CASCADE", "CASCADE").Error
				if err != nil {
					return err
				}

				return tx.Exec("ALTER TABLE test_results ALTER COLUMN test_no DROP DEFAULT").Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.DropTable(&submission.TestResult{}, &submission.Result{}, &submission.Submission{}).Error
			},
		},
		{
			// make test results' scores floating point
			ID: "201701200015",
			Migrate: func(tx *gorm.DB) error {
				return tx.Model(&submission.TestResult{}).ModifyColumn("score", "numeric(3,2)").Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Model(&submission.TestResult{}).ModifyColumn("score", "integer").Error
			},
		},
		{
			ID: "201701200030",
			Migrate: func(tx *gorm.DB) error {
				type SubmissionResult struct {
					Score decimal.Decimal `gorm:"type:numeric(5,2)"`
				}

				return tx.AutoMigrate(&SubmissionResult{}).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Model(&submission.Result{}).DropColumn("score").Error
			},
		},
		{
			ID: "201801300015",
			Migrate: func(tx *gorm.DB) error {
				return tx.Model(&submission.Submission{}).AddIndex("idx_submissions_created_at", "created_at").Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Model(&submission.Submission{}).RemoveIndex("idx_submissions_created_at").Error
			},
		},
		{
			// Submission results get a build command
			ID: "201802200015",
			Migrate: func(tx *gorm.DB) error {
				type SubmissionResult struct {
					BuildCommand string
				}
				return tx.AutoMigrate(&SubmissionResult{}).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Model(&submission.Result{}).DropColumn("build_command").Error
			},
		},
		{
			// Attachments have a size
			ID: "201802200030",
			Migrate: func(tx *gorm.DB) error {
				type Attachment struct {
					Size int32
				}
				return tx.AutoMigrate(&Attachment{}).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Model(&attachment.Attachment{}).DropColumn("size").Error
			},
		},
		{
			// Timestamps for attachments
			ID: "201802200045",
			Migrate: func(tx *gorm.DB) error {
				type Attachment struct {
					CreatedAt time.Time
					UpdatedAt time.Time
				}
				return tx.AutoMigrate(&Attachment{}).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Model(&attachment.Attachment{}).DropColumn("created_at").DropColumn("updated_at").Error
			},
		},
		{
			ID: "201802250015",
			Migrate: func(tx *gorm.DB) error {
				type Page struct {
					ID              uuid.UUID `gorm:"primary_key;type:uuid;default:uuid_generate_v1mc()"`
					Path            string    `gorm:"unique_index"`
					LatestTimestamp time.Time
					CreatedAt       time.Time
				}
				type PageVersion struct {
					PageID       uuid.UUID `gorm:"primary_key;type:uuid"`
					Timestamp    time.Time `gorm:"primary_key"`
					AttachmentID uuid.UUID `gorm:"type:uuid"`
				}

				if err := tx.AutoMigrate(&Page{}, &PageVersion{}).Error; err != nil {
					return err
				}

				return tx.Model(&PageVersion{}).AddForeignKey("page_id", "pages(id)", "CASCADE", "CASCADE").Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.DropTable(&page.Page{}, &page.Version{}).Error
			},
		},
		{
			ID: "201802250030",
			Migrate: func(tx *gorm.DB) error {
				type Page struct {
					DeletedAt *time.Time
				}
				return tx.AutoMigrate(&Page{}).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Model(&page.Version{}).DropColumn("deleted_at").Error
			},
		},
		{
			ID: "201802260015",
			Migrate: func(tx *gorm.DB) error {
				type PageVersion struct {
					DeletedAt *time.Time
				}
				return tx.AutoMigrate(&PageVersion{}).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Model(&page.Version{}).DropColumn("deleted_at").Error
			},
		},
		{
			ID: "201802280015",
			Migrate: func(tx *gorm.DB) error {
				type PageVersion struct {
					Title string
				}
				return tx.AutoMigrate(&PageVersion{}).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Model(&page.Version{}).DropColumn("title").Error
			},
		},
		{
			// give tasks a title
			ID: "201803140015",
			Migrate: func(tx *gorm.DB) error {
				type Task struct {
					Title  string
					PageID uuid.UUID `gorm:"type:uuid"`
				}
				return tx.AutoMigrate(&Task{}).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Model(&problem.Task{}).DropColumn("title").DropColumn("page_id").Error
			},
		},
		{
			ID: "201803140030",
			Migrate: func(tx *gorm.DB) error {
				return tx.Model(&problem.Task{}).AddForeignKey("page_id", "pages(id)", "RESTRICT", "CASCADE").Error
			},
			Rollback: func(tx *gorm.DB) error {
				return rfk(tx, "tasks", "page_id", "pages(id)")
			},
		},
		{
			ID: "201803150015",
			Migrate: func(tx *gorm.DB) error {
				type TaskList struct {
					ID          uuid.UUID `gorm:"primary_key;type:uuid;default:uuid_generate_v1mc()"`
					PageID      uuid.UUID `gorm:"type:uuid"`
					StartTime   time.Time
					EndTime     time.Time
					Name        string
					Description string
					Title       string
				}

				return tx.AutoMigrate(&TaskList{}).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.DropTable("task_lists").Error
			},
		},
		{
			ID: "201803180015",
			Migrate: func(tx *gorm.DB) error {
				type Task struct {
					TaskListID uuid.UUID `gorm:"type:uuid"`
				}

				err := tx.AutoMigrate(&Task{}).Error
				if err != nil {
					return err
				}
				return tx.Model(&Task{}).AddForeignKey("task_list_id", "task_lists(id)", "RESTRICT", "CASCADE").Error
			},
			Rollback: func(tx *gorm.DB) error {
				err := rfk(tx, "tasks", "task_list_id", "task_lists(id)")
				if err != nil {
					return err
				}
				return tx.Model(&problem.Task{}).DropColumn("task_list_id").Error
			},
		},
		{
			ID: "201803190015",
			Migrate: func(tx *gorm.DB) error {
				return tx.Model(&tasklist.TaskList{}).AddUniqueIndex("idx_task_lists_name", "name").Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Model(&tasklist.TaskList{}).RemoveIndex("idx_task_lists_name").Error
			},
		},
		{
			ID: "201803250015",
			Migrate: func(tx *gorm.DB) error {
				type Submission struct {
					UserID string `gorm:"type:uuid"`
				}
				return tx.AutoMigrate(&Submission{}).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Model(&submission.Submission{}).DropColumn("user_id").Error
			},
		},
		{
			ID: "201804060015",
			Migrate: func(tx *gorm.DB) error {
				return tx.Exec("ALTER TABLE tasks ALTER COLUMN page_id SET DEFAULT NULL").Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Exec("ALTER TABLE tasks ALTER COLUMN page_id SET DEFAULT ?", uuid.Nil).Error
			},
		},
		{ // make page paths ltrees
			ID: "201805310015",
			Migrate: func(tx *gorm.DB) error {
				t := tx.Begin()
				err := t.Exec("UPDATE pages SET path = replace(path, '/', '.')").Error
				if err != nil {
					t.Rollback()
					return err
				}

				err = t.Exec("ALTER TABLE pages ALTER COLUMN path TYPE ltree USING path::ltree").Error
				if err != nil {
					t.Rollback()
					return err
				}

				return t.Commit().Error
			},
			Rollback: func(tx *gorm.DB) error {
				t := tx.Begin()
				err := t.Exec("ALTER TABLE pages ALTER COLUMN path TYPE text").Error
				if err != nil {
					t.Rollback()
					return err
				}

				err = t.Exec("UPDATE pages SET path = replace(path, '.', '/')").Error
				if err != nil {
					t.Rollback()
					return err
				}

				return t.Commit().Error
			},
		},
		{
			ID: "201806080015",
			Migrate: func(tx *gorm.DB) error {
				type Page struct {
					ParentID uuid.UUID `gorm:"type:uuid"`
				}
				if err := tx.AutoMigrate(&Page{}).Error; err != nil {
					return err
				}
				return tx.Exec(xmcPageChildren["201806080015"]).Error
			},
			Rollback: func(tx *gorm.DB) error {
				if err := tx.Model(&page.Page{}).DropColumn("parent_id").Error; err != nil {
					return err
				}
				return tx.Exec("DROP FUNCTION xmc_page_children").Error
			},
		},
		{ // adds internals table that stores internals variables
			ID: "201806100015",
			Migrate: func(tx *gorm.DB) error {
				type Internal struct {
					ID                     uint `gorm:"primary_key;auto_increment:false;default:0"`
					LastPageEvent          time.Time
					LastPageChildrenUpdate time.Time
				}
				err := tx.Table("internal").CreateTable(&Internal{}).Error
				if err != nil {
					return err
				}
				return tx.Table("internal").Create(&Internal{}).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.DropTable("internal").Error
			},
		},
		{
			ID: "201806100030",
			Migrate: func(tx *gorm.DB) error {
				return tx.Exec(xmcPageChildren["201806100030"]).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Exec(xmcPageChildren["201806080015"]).Error
			},
		},
		{ // new field for attachments, is_public
			ID: "201806140015",
			Migrate: func(tx *gorm.DB) error {
				type Attachment struct {
					IsPublic bool
				}
				return tx.AutoMigrate(&Attachment{}).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Model(&attachment.Attachment{}).DropColumn("is_public").Error
			},
		},
		{
			ID: "201807130015",
			Migrate: func(tx *gorm.DB) error {
				type PageVersion struct {
					Contents string `gorm:"type:text"`
				}
				if err := tx.AutoMigrate(&PageVersion{}).Error; err != nil {
					return err
				}

				return tx.Model(&PageVersion{}).DropColumn("attachment_id").Error
			},
			Rollback: func(tx *gorm.DB) error {
				type PageVersion struct {
					AttachmentID uuid.UUID `gorm:"type:uuid"`
				}
				if err := tx.Model(&page.Version{}).DropColumn("contents").Error; err != nil {
					return err
				}

				return tx.AutoMigrate(&PageVersion{}).Error
			},
		},
		{
			ID: "201807150015",
			Migrate: func(tx *gorm.DB) error {
				type TaskList struct {
					PublicSubmissions bool
				}
				return tx.AutoMigrate(&TaskList{}).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Model(&tasklist.TaskList{}).DropColumn("public_submissions").Error
			},
		},
		{
			ID: "201807200015",
			Migrate: func(tx *gorm.DB) error {
				type TaskList struct {
					WithParticipations bool
				}
				type Participation struct {
					TaskListID uuid.UUID `gorm:"primary_key;type:uuid"`
					UserID     uuid.UUID `gorm:"primary_key"`
				}
				if err := tx.AutoMigrate(&TaskList{}).AutoMigrate(&Participation{}).Error; err != nil {
					return err
				}

				err := tx.Model(&Participation{}).
					AddForeignKey("task_list_id", "task_lists(id)", "CASCADE", "CASCADE").
					Error
				return err
			},
			Rollback: func(tx *gorm.DB) error {
				if err := tx.Model(&tasklist.TaskList{}).DropColumn("with_participations").Error; err != nil {
					return err
				}

				return tx.DropTable(&tasklist.Participation{}).Error
			},
		},
		{
			ID: "201807240015",
			Migrate: func(tx *gorm.DB) error {
				type Page struct {
					ObjectID string
				}

				return tx.AutoMigrate(&Page{}).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Model(&page.Page{}).DropColumn("object_id").Error
			},
		},
	})
	return errors.Wrap(m.Migrate(), "failed to migrate schema")
}
