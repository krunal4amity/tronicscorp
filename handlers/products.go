package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/krunal4amity/tronicscorp/dbiface"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//Product describes an electronic product e.g. phone
type Product struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name        string             `json:"product_name" bson:"product_name" validate:"required,max=10"`
	Price       int                `json:"price" bson:"price" validate:"required,max=2000"`
	Currency    string             `json:"currency" bson:"currency" validate:"required,len=3"`
	Discount    int                `json:"discount" bson:"discount"`
	Vendor      string             `json:"vendor" bson:"vendor" validate:"required"`
	Accessories []string           `json:"accessories,omitempty" bson:"accessories,omitempty"`
	IsEssential bool               `json:"is_essential" bson:"is_essential"`
}

//ProductHandler a product handler
type ProductHandler struct {
	Col dbiface.CollectionAPI
}

func findProducts(ctx context.Context, q url.Values, collection dbiface.CollectionAPI) ([]Product, *echo.HTTPError) {
	var products []Product
	filter := make(map[string]interface{})
	for k, v := range q {
		filter[k] = v[0]
	}
	if filter["_id"] != nil {
		docID, err := primitive.ObjectIDFromHex(filter["_id"].(string))
		if err != nil {
			log.Errorf("Unable to convert to Object ID : %v", err)
			return products,
				echo.NewHTTPError(http.StatusInternalServerError, errorMessage{Message: "unable to convert to ObjectID"})
		}
		filter["_id"] = docID
	}
	cursor, err := collection.Find(ctx, bson.M(filter))
	if err != nil {
		log.Errorf("Unable to find the products : %v", err)
		return products,
			echo.NewHTTPError(http.StatusNotFound, errorMessage{Message: "unable to find the products"})
	}
	err = cursor.All(ctx, &products)
	if err != nil {
		log.Errorf("Unable to read the cursor : %v", err)
		return products,
			echo.NewHTTPError(http.StatusUnprocessableEntity, errorMessage{Message: "unable to parse retrieved products"})
	}
	return products, nil
}

//GetProducts gets a list of products
func (h *ProductHandler) GetProducts(c echo.Context) error {
	products, httpError := findProducts(context.Background(), c.QueryParams(), h.Col)
	if httpError != nil {
		return c.JSON(httpError.Code, httpError.Message)
	}
	return c.JSON(http.StatusOK, products)
}

func findProduct(ctx context.Context, id string, collection dbiface.CollectionAPI) (Product, *echo.HTTPError) {
	var product Product
	docID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Errorf("Unable to convert to Object ID : %v", err)
		return product,
			echo.NewHTTPError(http.StatusInternalServerError, errorMessage{Message: "unable to convert to ObjectID"})
	}
	res := collection.FindOne(ctx, bson.M{"_id": docID})
	err = res.Decode(&product)
	if err != nil {
		log.Errorf("Unable to find the product : %v", err)
		return product,
			echo.NewHTTPError(http.StatusNotFound, errorMessage{Message: "unable to find the product"})
	}
	return product, nil
}

//GetProduct gets a single product
func (h *ProductHandler) GetProduct(c echo.Context) error {
	product, httpError := findProduct(context.Background(), c.Param("id"), h.Col)
	if httpError != nil {
		return c.JSON(httpError.Code, httpError.Message)
	}
	return c.JSON(http.StatusOK, product)
}

func deleteProduct(ctx context.Context, id string, collection dbiface.CollectionAPI) (int64, *echo.HTTPError) {
	docID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Errorf("Unable convert to ObjectID : %v", err)
		return 0,
			echo.NewHTTPError(http.StatusInternalServerError, errorMessage{Message: "unable to convert to ObjectID"})
	}
	res, err := collection.DeleteOne(ctx, bson.M{"_id": docID})
	if err != nil {
		log.Errorf("Unable to delete the product : %v", err)
		return 0,
			echo.NewHTTPError(http.StatusInternalServerError, errorMessage{Message: "unable to delete the product"})
	}
	return res.DeletedCount, nil
}

//DeleteProduct gets a single product
func (h *ProductHandler) DeleteProduct(c echo.Context) error {
	delCount, httpError := deleteProduct(context.Background(), c.Param("id"), h.Col)
	if httpError != nil {
		return c.JSON(httpError.Code, httpError.Message)
	}
	return c.JSON(http.StatusOK, delCount)
}

func modifyProduct(ctx context.Context, id string, reqBody io.ReadCloser, collection dbiface.CollectionAPI) (Product, *echo.HTTPError) {
	var product Product
	//find if the product exits, if err return 404
	docID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Errorf("cannot convert to ObjectID :%v", err)
		return product,
			echo.NewHTTPError(http.StatusInternalServerError, errorMessage{Message: "unable to convert to ObjectID"})
	}
	filter := bson.M{"_id": docID}
	res := collection.FindOne(ctx, filter)
	if err := res.Decode(&product); err != nil {
		log.Errorf("unable to decode to product :%v", err)
		return product,
			echo.NewHTTPError(http.StatusUnprocessableEntity, errorMessage{Message: "unable to find the product"})
	}

	//decode the req payload, if err return 500
	if err := json.NewDecoder(reqBody).Decode(&product); err != nil {
		log.Errorf("unable to decode using reqbody : %v", err)
		return product,
			echo.NewHTTPError(http.StatusUnprocessableEntity, errorMessage{Message: "unable to parse request payload"})
	}

	//validate the request, if err return 400
	if err := v.Struct(product); err != nil {
		log.Errorf("unable to validate the struct : %v", err)
		return product,
			echo.NewHTTPError(http.StatusBadRequest, errorMessage{Message: "unable to validate the request payload"})
	}

	//update the product, if err return 500
	_, err = collection.UpdateOne(ctx, filter, bson.M{"$set": product})
	if err != nil {
		log.Errorf("Unable to update the product : %v", err)
		return product,
			echo.NewHTTPError(http.StatusInternalServerError, errorMessage{Message: "unable to update the product"})
	}
	return product, nil
}

//UpdateProduct updates a product
func (h *ProductHandler) UpdateProduct(c echo.Context) error {
	product, httpError := modifyProduct(context.Background(), c.Param("id"), c.Request().Body, h.Col)
	if httpError != nil {
		return c.JSON(httpError.Code, httpError.Message)
	}
	return c.JSON(http.StatusOK, product)
}

func insertProducts(ctx context.Context, products []Product, collection dbiface.CollectionAPI) ([]interface{}, *echo.HTTPError) {
	var insertedIds []interface{}
	for _, product := range products {
		product.ID = primitive.NewObjectID()
		insertID, err := collection.InsertOne(ctx, product)
		if err != nil {
			log.Errorf("Unable to insert to Database:%v", err)
			return nil,
				echo.NewHTTPError(http.StatusInternalServerError, errorMessage{Message: "unable to insert to database"})
		}
		insertedIds = append(insertedIds, insertID.InsertedID)
	}
	return insertedIds, nil
}

//CreateProducts create products on mongodb database
func (h *ProductHandler) CreateProducts(c echo.Context) error {
	var products []Product
	c.Echo().Validator = &ProductValidator{validator: v}
	if err := c.Bind(&products); err != nil {
		log.Errorf("Unable to bind : %v", err)
		return c.JSON(http.StatusUnprocessableEntity, errorMessage{Message: "unable to parse request payload"})
	}
	for _, product := range products {
		if err := c.Validate(product); err != nil {
			log.Errorf("Unable to validate the product %+v %v", product, err)
			return c.JSON(http.StatusBadRequest, errorMessage{Message: "unable to validate request payload"})
		}
	}
	IDs, httpError := insertProducts(context.Background(), products, h.Col)
	if httpError != nil {
		return c.JSON(httpError.Code, httpError.Message)
	}
	return c.JSON(http.StatusCreated, IDs)
}
